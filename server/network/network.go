package network

import (
	"errors"
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server/client"
	"github.com/maria-mz/bash-battle-server/server/event_builder"
)

var ErrUsernameTaken = errors.New("a player with this name already exists")
var ErrClientNotFound = errors.New("client not found in game")

type ClientMsg struct {
	Username string
	Msg      *pb.AckMsg
}

type Network struct {
	clients        map[string]*client.Client
	clientMsgs     chan<- ClientMsg
	activityBitmap ActivityBitmap
}

func NewNetwork() (*Network, <-chan ClientMsg) {
	clients := make(map[string]*client.Client)
	bitmap := NewActivityBitmap()
	ch := make(chan ClientMsg)

	net := &Network{
		clients:        clients,
		activityBitmap: bitmap,
		clientMsgs:     ch,
	}

	return net, ch
}

func (net *Network) AddClient(client *client.Client) error {
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

	msg := <-client.Stream.EndStreamMsgs // blocking

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

func (net *Network) handleClientMsgs(client *client.Client) {
	for msg := range client.Stream.AckMsgs {
		client.Active = true

		switch msg.Ack.(type) {

		case *pb.AckMsg_RoundLoaded:
			log.Logger.Info(
				"Client loaded round [ACK]", "client", client.Username,
			)
			net.activityBitmap.SetLoadAck(client.Username, true)

		case *pb.AckMsg_RoundSubmission:
			log.Logger.Info(
				"Client made a submission [ACK]", "client", client.Username,
			)
			net.activityBitmap.SetSubmissionAck(client.Username, true)
		}

		net.clientMsgs <- ClientMsg{client.Username, msg}
	}
}

func (net *Network) BroadcastPlayerJoin(player *game.Player) {
	log.Logger.Info(
		"Broadcasting event PLAYER_JOIN", "player", player.InfoString(),
	)

	event := event_builder.BuildPlayerJoinedEvent(player)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastCountdown(round int, startsAt time.Time) {
	log.Logger.Info(
		"Broadcasting event COUNTING_DOWN", "round", round, "startsAt", startsAt.UTC(),
	)

	event := event_builder.BuildCountingDownEvent(round, startsAt)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastRoundStart(round int, endsAt time.Time) {
	log.Logger.Info(
		"Broadcasting event ROUND_STARTED", "round", round, "endsAt", endsAt.UTC(),
	)

	event := event_builder.BuildRoundStartedEvent(round, endsAt)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastGameOver() {
	log.Logger.Info("Broadcasting event GAME_OVER")

	event := event_builder.BuildGameOverEvent()
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastLoadRoundCmd(round int, challenge game.Challenge) {
	log.Logger.Info(
		"Broadcasting command LOAD_ROUND", "round", round, "challenge", challenge.InfoString(),
	)

	cmd := event_builder.BuildLoadRoundEvent(round, challenge)
	net.BroadcastEvent(cmd)
}

func (net *Network) BroadcastSubmitScoreCmd(round int) {
	log.Logger.Info("Broadcasting command SUBMIT_ROUND_SCORE", "round", round)

	cmd := event_builder.BuildSubmitRoundScoreEvent()
	net.BroadcastEvent(cmd)
}

func (net *Network) BroadcastEvent(event *pb.Event) {
	for _, client := range net.clients {
		net.sendEventToClient(event, client)
	}
}

func (net *Network) sendEventToClient(event *pb.Event, client *client.Client) {
	if client.Stream != nil {
		log.Logger.Info(
			"Sent event to client", "client", client.Username,
		)
		client.Stream.SendEvent(event)
	} else {
		log.Logger.Info(
			"Did not send event to client (stream is nil)", "client", client.Username,
		)
	}
}

func (net *Network) AllClientsLoaded() bool {
	return net.activityBitmap.CountAcks(AckLoad) == len(net.clients)
}

func (net *Network) AllClientsSubmitted() bool {
	return net.activityBitmap.CountAcks(AckSubmission) == len(net.clients)
}

func (net *Network) GetClientLoadStatus(username string) bool {
	return net.activityBitmap.GetStatus(AckLoad, username)
}

func (net *Network) GetClientSubmissionStatus(username string) bool {
	return net.activityBitmap.GetStatus(AckSubmission, username)
}

func (net *Network) ResetAcks() {
	net.activityBitmap.ResetAcks()
}
