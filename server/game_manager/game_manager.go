package game_manager

import (
	"errors"
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server/network"
)

var ErrJoinOnGameStarted = errors.New("cannot join game: game already started")
var ErrStreamOnGameOver = errors.New("cannot stream game: game is over")

type state int

const (
	Lobby state = iota
	Load
	Play
	Submission
	Done
	Terminated
)

type GameManager struct {
	network    *network.Network
	clientMsgs <-chan network.ClientMsg

	gameData         *game.GameData
	gameRunner       *game.GameRunner
	gameRunnerEvents <-chan game.RunnerEvent

	state state
}

func NewGameManager(config config.GameConfig) *GameManager {
	broadcaster, clientMsgs := network.NewNetwork()
	gameData := game.NewGameData(config)
	gameRunner, gameRunnerEvents := game.NewGameRunner(gameData)

	gm := &GameManager{
		network:          broadcaster,
		clientMsgs:       clientMsgs,
		gameData:         gameData,
		gameRunner:       gameRunner,
		gameRunnerEvents: gameRunnerEvents,
	}

	go gm.handleRunnerEvents()
	go gm.handleClientMsgs()

	return gm
}

func (gm *GameManager) handleRunnerEvents() {
	for event := range gm.gameRunnerEvents {
		round := gm.gameRunner.GetCurrentRound()

		switch event {
		case game.CountingDown:
			gm.onCountingDown(round)

		case game.RoundStarted:
			gm.onRoundStarted(round)

		case game.RoundEnded:
			gm.onRoundEnded(round)
		}
	}

	log.Logger.Info("exiting loop!!!!!!\n")
}

func (gm *GameManager) onCountingDown(round int) {
	roundStartsAt := time.Now().Add(gm.gameData.GetCountdownDuration())
	go gm.network.BroadcastCountdown(round, roundStartsAt)
}

func (gm *GameManager) onRoundStarted(round int) {
	roundEndsAt := time.Now().Add(gm.gameData.GetRoundDuration())
	go gm.network.BroadcastRoundStart(round, roundEndsAt)
}

func (gm *GameManager) onRoundEnded(round int) {
	gm.state = Submission
	go gm.network.BroadcastSubmitScore(round, gm.onSubmitScoreBroadcasted)
}

func (gm *GameManager) onSubmitScoreBroadcasted() {
	if gm.gameRunner.IsFinalRound() {
		gm.state = Done
		gm.network.BroadcastGameOver()
	} else {
		gm.state = Load
		gm.loadNextRound()
	}
}

func (gm *GameManager) onLoadRoundBroadcasted() {
	gm.state = Play
	gm.gameRunner.RunRound()
}

func (gm *GameManager) AddClient(client *network.Client) error {
	if gm.state != Lobby {
		return ErrJoinOnGameStarted
	}

	err := gm.network.AddClient(client)
	if err != nil {
		return err
	}

	player := game.NewPlayer(client.Username)

	gm.gameData.AddPlayer(player)
	gm.network.BroadcastPlayerJoin(player)

	if gm.gameData.IsGameFull() {
		gm.state = Load
		gm.loadNextRound()
	}

	return nil
}

func (gm *GameManager) GetPlayers() []*game.Player {
	return gm.gameData.GetPlayers()
}

func (gm *GameManager) ListenForClientMsgs(client *network.Client) error {
	if gm.state == Done {
		return ErrStreamOnGameOver
	}
	err := gm.network.ListenForClientMsgs(client.Username) // Blocking
	return err
}

func (gm *GameManager) handleClientMsgs() {
	for msg := range gm.clientMsgs {
		switch ack := msg.Msg.GetAck().(type) {

		case *pb.AckMsg_RoundLoaded:
			continue // do nothing

		case *pb.AckMsg_RoundSubmission:
			if gm.state == Submission {
				gm.makeSubmission(ack.RoundSubmission.RoundStats, msg.Username)
			}
		}
	}
}

func (gm *GameManager) makeSubmission(stats *pb.RoundStats, username string) {
	score := game.Score{
		Round: gm.gameRunner.GetCurrentRound(),
		Win:   stats.Won,
	}

	player, ok := gm.gameData.GetPlayer(username)

	if !ok {
		log.Logger.Fatal("failed to set score for player", "username", username)
	}
	player.SetRoundScore(score)
}

func (gm *GameManager) loadNextRound() {
	round := gm.gameRunner.GetCurrentRound() + 1
	challenge, ok := gm.gameData.GetChallenge(round - 1) // 0-based

	if !ok {
		log.Logger.Fatal("No challenge found for round", "round", round)
	}

	go gm.network.BroadcastLoadRound(round, challenge, gm.onLoadRoundBroadcasted)
}
