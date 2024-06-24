package server

import (
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type state int

const (
	Lobby state = iota
	Load
	Play
	Submission
	Done
)

type gameManager struct {
	clients         map[string]*client // Subset of server's Client registry
	clientAckBitmap *ClientAckStatusBitmap

	gameData         *game.GameData
	gameRunner       *game.GameRunner
	gameRunnerEvents <-chan game.RunnerEvent

	state state
}

func NewGameManager(config config.GameConfig) *gameManager {
	clients := make(map[string]*client)
	clientAckBitmap := NewClientAckBitmap()
	gameData := game.NewGameData(config)
	gameRunner, gameRunnerEvents := game.NewGameRunner(gameData)

	manager := &gameManager{
		clients:          clients,
		clientAckBitmap:  &clientAckBitmap,
		gameData:         gameData,
		gameRunner:       gameRunner,
		gameRunnerEvents: gameRunnerEvents,
	}

	go manager.handleRunnerEvents()

	return manager
}

func (manager *gameManager) handleRunnerEvents() {
	for event := range manager.gameRunnerEvents {
		round := manager.gameRunner.GetCurrentRound()

		switch event {
		case game.CountingDown:
			roundStartsAt := time.Now().Add(manager.gameData.GetCountdownDuration())
			manager.broadcastCountdown(round, roundStartsAt)

		case game.RoundStarted:
			roundEndsAt := time.Now().Add(manager.gameData.GetRoundDuration())
			manager.broadcastRoundStart(round, roundEndsAt)

		case game.RoundEnded:
			manager.state = Submission
			manager.requestClientsToSubmitScores()

		case game.GameDone:
			return
		}
	}
}

func (manager *gameManager) handleClientAckMsgs(client *client) {
	for msg := range client.stream.ackMsgs {
		switch msg.Ack.(type) {
		case *pb.AckMsg_RoundLoaded:
			manager.handleLoadAck(client)

		case *pb.AckMsg_RoundSubmission:
			manager.handleSubmissionAck(msg, client)
		}
	}
}

func (manager *gameManager) ListenToClientStream(client *client) error {
	if manager.state == Done {
		return ErrStreamOnGameOver
	}

	client, ok := manager.clients[client.username]
	if !ok {
		return ErrClientNotFound
	}

	go manager.handleClientAckMsgs(client)
	go client.stream.Recv()

	msg := <-client.stream.endStreamMsgs // blocking

	if msg.err != nil {
		log.Logger.Warn(
			"Stream ended due to error", "client", client, "err", msg.err,
		)

		return msg.err
	} else {
		log.Logger.Info(
			"Stream ended gracefully", "client", client, "info", msg.info,
		)

		return nil
	}
}

func (manager *gameManager) handleLoadAck(client *client) {
	if manager.state != Load {
		return
	}

	manager.clientAckBitmap.SetLoadAck(client.username, true)
	manager.checkLoads()
}

func (manager *gameManager) checkLoads() {
	allLoaded := manager.clientAckBitmap.CountAcks(AckLoad) == len(manager.clients)

	if allLoaded {
		manager.state = Play
		manager.gameRunner.RunRound()
	}
}

func (manager *gameManager) handleSubmissionAck(ack *pb.AckMsg, client *client) {
	if manager.state != Submission {
		return
	}

	submissionMsg := ack.GetRoundSubmission()

	score := game.Score{
		Round: manager.gameRunner.GetCurrentRound(),
		Win:   submissionMsg.RoundStats.Won,
	}

	player, _ := manager.gameData.GetPlayer(client.username)
	player.SetRoundScore(score)

	manager.clientAckBitmap.SetSubmissionAck(client.username, true)
	manager.checkSubmissions()
}

func (manager *gameManager) checkSubmissions() {
	allSubmitted := manager.clientAckBitmap.CountAcks(AckSubmission) == len(manager.clients)

	if !allSubmitted {
		return
	}

	manager.clientAckBitmap.ResetAcks()

	if manager.gameRunner.IsFinalRound() {
		manager.state = Done
		manager.broadcastGameOver()
	} else {
		manager.state = Load
		manager.requestClientsToLoadRound(manager.gameRunner.GetCurrentRound() + 1)
	}
}

func (manager *gameManager) AddClient(client *client) error {
	if manager.state != Lobby {
		return ErrJoinOnGameStarted
	}

	_, ok := manager.clients[client.username]
	if ok {
		return ErrUsernameTaken
	}

	manager.clients[client.username] = client

	player := game.NewPlayer(client.username)

	manager.gameData.AddPlayer(player)
	manager.broadcastPlayerJoin(player)

	if manager.gameData.IsGameFull() {
		manager.state = Load
		manager.requestClientsToLoadRound(1)
	}

	return nil
}

func (manager *gameManager) broadcastPlayerJoin(player *game.Player) {
	event := &pb.Event{
		Event: &pb.Event_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{Player: player.ToProto()},
		},
	}

	log.Logger.Debug("Broadcasting 'PlayerJoined' event", "event", event)
	manager.broadcastEvent(event)
}

func (manager *gameManager) requestClientsToLoadRound(round int) {
	challenge, ok := manager.gameData.GetChallenge(round)
	if !ok {
		return
	}

	event := &pb.Event{
		Event: &pb.Event_LoadRound{
			LoadRound: &pb.LoadRound{
				RoundNumber: int32(round),
				Question:    challenge.Question,
				// TODO: Add files as bytes
			},
		},
	}

	log.Logger.Debug("Broadcasting 'LoadRound' event", "event", event)
	manager.broadcastEvent(event)
}

func (manager *gameManager) broadcastCountdown(round int, startsAt time.Time) {
	event := &pb.Event{
		Event: &pb.Event_CountingDown{
			CountingDown: &pb.CountingDown{
				RoundNumber: int32(round),
				StartsAt:    timestamppb.New(startsAt),
			},
		},
	}

	log.Logger.Debug("Broadcasting 'CountingDown' event", "event", event)
	manager.broadcastEvent(event)
}

func (manager *gameManager) broadcastRoundStart(round int, endsAt time.Time) {
	event := &pb.Event{
		Event: &pb.Event_RoundStarted{
			RoundStarted: &pb.RoundStarted{
				RoundNumber: int32(round),
				EndsAt:      timestamppb.New(endsAt),
			},
		},
	}

	log.Logger.Debug("Broadcasting 'RoundStarted' event", "event", event)
	manager.broadcastEvent(event)
}

func (manager *gameManager) requestClientsToSubmitScores() {
	event := &pb.Event{
		Event: &pb.Event_SubmitRoundScore{},
	}

	log.Logger.Debug("Broadcasting 'SubmitRoundScore' event", "event", event)
	manager.broadcastEvent(event)
}

func (manager *gameManager) broadcastGameOver() {
	event := &pb.Event{
		Event: &pb.Event_GameOver{
			GameOver: &pb.GameOver{},
		},
	}

	log.Logger.Debug("Broadcasting 'GameOver' event", "event", event)
	manager.broadcastEvent(event)
}

func (manager *gameManager) broadcastEvent(event *pb.Event) {
	for _, client := range manager.clients {
		if client.stream != nil {
			log.Logger.Debug("Sending event to client", "client", client)
			client.stream.SendEvent(event)
		}
	}
}
