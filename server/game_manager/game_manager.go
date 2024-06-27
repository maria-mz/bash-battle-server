package game_manager

import (
	"errors"
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game"
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
	network, clientMsgs := network.NewNetwork()
	gameData := game.NewGameData(config)
	gameRunner, gameRunnerEvents := game.NewGameRunner(gameData)

	manager := &GameManager{
		network:          network,
		clientMsgs:       clientMsgs,
		gameData:         gameData,
		gameRunner:       gameRunner,
		gameRunnerEvents: gameRunnerEvents,
	}

	go manager.handleRunnerEvents()
	go manager.handleClientMsgs()

	return manager
}

func (manager *GameManager) handleRunnerEvents() {
	for event := range manager.gameRunnerEvents {
		round := manager.gameRunner.GetCurrentRound()

		switch event {
		case game.CountingDown:
			roundStartsAt := time.Now().Add(manager.gameData.GetCountdownDuration())
			manager.network.BroadcastCountdown(round, roundStartsAt)

		case game.RoundStarted:
			roundEndsAt := time.Now().Add(manager.gameData.GetRoundDuration())
			manager.network.BroadcastRoundStart(round, roundEndsAt)

		case game.RoundEnded:
			manager.state = Submission
			manager.network.BroadcastSubmitScoreCmd(round)

		case game.GameDone:
			return
		}
	}
}

func (manager *GameManager) AddClient(client *network.Client) error {
	if manager.state != Lobby {
		return ErrJoinOnGameStarted
	}

	if err := manager.network.AddClient(client); err != nil {
		return err
	}

	player := game.NewPlayer(client.Username)

	manager.gameData.AddPlayer(player)
	manager.network.BroadcastPlayerJoin(player)

	if manager.gameData.IsGameFull() {
		manager.state = Load
		manager.broadcastLoadNextRoundCmd()
	}

	return nil
}

func (manager *GameManager) ListenForClientMsgs(username string) error {
	err := manager.network.ListenForClientMsgs(username) // Blocking
	return err
}

func (manager *GameManager) handleClientMsgs() {
	for msg := range manager.clientMsgs {
		switch ack := msg.Msg.GetAck().(type) {

		case *pb.AckMsg_RoundLoaded:
			if manager.state != Load {
				continue
			}
			manager.checkLoads()

		case *pb.AckMsg_RoundSubmission:
			if manager.state != Submission {
				return
			}
			manager.makeSubmission(ack, msg.Username)
			manager.checkSubmissions()
		}
	}
}

func (manager *GameManager) makeSubmission(msg *pb.AckMsg_RoundSubmission, username string) {
	score := game.Score{
		Round: manager.gameRunner.GetCurrentRound(),
		Win:   msg.RoundSubmission.RoundStats.Won,
	}

	player, _ := manager.gameData.GetPlayer(username)
	player.SetRoundScore(score)
}

func (manager *GameManager) checkLoads() {
	if manager.network.AllClientsLoaded() {
		manager.state = Play
		manager.gameRunner.RunRound()
	}
}

func (manager *GameManager) checkSubmissions() {
	if !manager.network.AllClientsSubmitted() {
		return
	}

	manager.network.ResetAcks()

	if manager.gameRunner.IsFinalRound() {
		manager.state = Done
		manager.network.BroadcastGameOver()
	} else {
		manager.state = Load
		manager.broadcastLoadNextRoundCmd()
	}
}

func (manager *GameManager) broadcastLoadNextRoundCmd() {
	round := manager.gameRunner.GetCurrentRound() + 1

	challenge, ok := manager.gameData.GetChallenge(round)
	if ok {
		manager.network.BroadcastLoadRoundCmd(round, challenge)
	}
}
