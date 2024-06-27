package server

import (
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game"
)

type state int

const (
	Lobby state = iota
	Load
	Play
	Submission
	Done
	Terminated
)

type gameManager struct {
	network    *Network
	clientMsgs <-chan ClientMsg

	gameData         *game.GameData
	gameRunner       *game.GameRunner
	gameRunnerEvents <-chan game.RunnerEvent

	state state
}

func NewGameManager(config config.GameConfig) *gameManager {
	network, clientMsgs := NewNetwork()
	gameData := game.NewGameData(config)
	gameRunner, gameRunnerEvents := game.NewGameRunner(gameData)

	manager := &gameManager{
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

func (manager *gameManager) handleRunnerEvents() {
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

func (manager *gameManager) AddClient(client *client) error {
	if manager.state != Lobby {
		return ErrJoinOnGameStarted
	}

	if err := manager.network.AddClient(client); err != nil {
		return err
	}

	player := game.NewPlayer(client.username)

	manager.gameData.AddPlayer(player)
	manager.network.BroadcastPlayerJoin(player)

	if manager.gameData.IsGameFull() {
		manager.state = Load
		manager.broadcastLoadNextRoundCmd()
	}

	return nil
}

func (manager *gameManager) ListenForClientMsgs(username string) error {
	err := manager.network.ListenForClientMsgs(username) // Blocking
	return err
}

func (manager *gameManager) handleClientMsgs() {
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

func (manager *gameManager) makeSubmission(msg *pb.AckMsg_RoundSubmission, username string) {
	score := game.Score{
		Round: manager.gameRunner.GetCurrentRound(),
		Win:   msg.RoundSubmission.RoundStats.Won,
	}

	player, _ := manager.gameData.GetPlayer(username)
	player.SetRoundScore(score)
}

func (manager *gameManager) checkLoads() {
	if manager.network.AllClientsLoaded() {
		manager.state = Play
		manager.gameRunner.RunRound()
	}
}

func (manager *gameManager) checkSubmissions() {
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

func (manager *gameManager) broadcastLoadNextRoundCmd() {
	round := manager.gameRunner.GetCurrentRound() + 1

	challenge, ok := manager.gameData.GetChallenge(round)
	if ok {
		manager.network.BroadcastLoadRoundCmd(round, challenge)
	}
}
