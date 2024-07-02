package network

import (
	"errors"
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
)

var ErrUsernameTaken = errors.New("a player with this name already exists")
var ErrClientNotFound = errors.New("client not found in game")

type ClientMsg struct {
	Username string
	Msg      *pb.AckMsg
}

type Network struct {
	clients    map[string]*Client
	clientMsgs chan<- ClientMsg
}

func NewNetwork() (*Network, <-chan ClientMsg) {
	clients := make(map[string]*Client)
	clientMsgs := make(chan ClientMsg)

	net := &Network{
		clients:    clients,
		clientMsgs: clientMsgs,
	}

	return net, clientMsgs
}

func (net *Network) AddClient(client *Client) error {
	if _, ok := net.clients[client.Username]; ok {
		return ErrUsernameTaken
	} else {
		net.clients[client.Username] = client
		return nil
	}
}

func (net *Network) ListenForClientMsgs(username string) error {
	client, ok := net.clients[username]

	if !ok {
		return ErrClientNotFound
	}

	go net.handleClientMsgs(client)
	go client.Stream.Recv()

	client.meta.Active = true

	msg := <-client.Stream.EndStreamMsgs // blocking

	client.meta.Active = false

	if msg.Err != nil {
		log.Logger.Warn(
			"Stream ended due to error", "client", client, "err", msg.Err,
		)
	} else {
		log.Logger.Info(
			"Stream ended gracefully", "client", client, "info", msg.Info,
		)
	}

	return msg.Err
}

func (net *Network) handleClientMsgs(client *Client) {
	for msg := range client.Stream.AckMsgs {
		switch msg.Ack.(type) {

		case *pb.AckMsg_RoundLoaded:
			log.Logger.Info("Client loaded round", "client", client.Username)

		case *pb.AckMsg_RoundSubmission:
			log.Logger.Info("Client made a submission", "client", client.Username)
		}

		net.clientMsgs <- ClientMsg{client.Username, msg}
	}
}

func (net *Network) BroadcastPlayerJoin(player *game.Player) {
	log.Logger.Info(
		"Broadcasting event PLAYER_JOIN", "player", player.InfoString(),
	)

	event := BuildPlayerJoinedEvent(player)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastCountdown(round int, startsAt time.Time) {
	log.Logger.Info(
		"Broadcasting event COUNTING_DOWN",
		"round", round,
		"startsAt", startsAt.UTC(),
	)

	event := BuildCountingDownEvent(round, startsAt)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastRoundStart(round int, endsAt time.Time) {
	log.Logger.Info(
		"Broadcasting event ROUND_STARTED",
		"round", round,
		"endsAt", endsAt.UTC(),
	)

	event := BuildRoundStartedEvent(round, endsAt)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastGameOver() {
	log.Logger.Info("Broadcasting event GAME_OVER")

	event := BuildGameOverEvent()
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastLoadRound(round int, challenge game.Challenge, callback func()) {
	net.broadcastMultipleTimes(
		func() { net.broadcastLoadRound(round, challenge) },
	)

	callback()
}

func (net *Network) BroadcastSubmitScore(round int, callback func()) {
	net.broadcastMultipleTimes(
		func() { net.broadcastSubmitScore(round) },
	)

	callback()
}

func (net *Network) broadcastLoadRound(round int, challenge game.Challenge) {
	log.Logger.Info(
		"Broadcasting event LOAD_ROUND",
		"round", round,
		"challenge", challenge.InfoString(),
	)

	event := BuildLoadRoundEvent(round, challenge)
	net.BroadcastEvent(event)
}

func (net *Network) broadcastSubmitScore(round int) {
	log.Logger.Info("Broadcasting event SUBMIT_ROUND_SCORE", "round", round)

	event := BuildSubmitRoundScoreEvent()
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastEvent(event *pb.Event) {
	for _, client := range net.clients {
		net.SendEventToClient(event, client)
	}
}

func (net *Network) broadcastMultipleTimes(broadcast func()) {
	ticker := time.NewTicker(1 * time.Second) // TODO: Use constant
	count := 0

	for range ticker.C {
		if count == 3 { // TODO: Use constant
			return
		}

		broadcast()
		count++
	}
}

func (net *Network) SendEventToClient(event *pb.Event, client *Client) {
	if client.meta.Active {
		log.Logger.Info(
			"Sent event to client", "client", client.Username,
		)
		client.Stream.SendEvent(event)
	} else {
		log.Logger.Info(
			"Did not send event to client (stream is nil)",
			"client", client.Username,
		)
	}
}
