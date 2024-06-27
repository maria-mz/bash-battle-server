package server

import (
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
)

type ClientMsg struct {
	Username string
	Msg      *pb.AckMsg
}

type Network struct {
	clients        map[string]*client
	clientMsgs     chan<- ClientMsg
	activityBitmap ActivityBitmap
}

func NewNetwork() (*Network, <-chan ClientMsg) {
	clients := make(map[string]*client)
	bitmap := NewActivityBitmap()
	ch := make(chan ClientMsg)

	net := &Network{
		clients:        clients,
		activityBitmap: bitmap,
		clientMsgs:     ch,
	}

	return net, ch
}

func (net *Network) AddClient(client *client) error {
	if _, ok := net.clients[client.username]; ok {
		return ErrUsernameTaken
	} else {
		net.clients[client.username] = client
		return nil
	}
}

func (net *Network) ListenForClientMsgs(username string) error {
	client, ok := net.clients[username]
	if !ok {
		return ErrClientNotFound
	}

	go net.handleClientMsgs(client)
	go client.stream.Recv()

	msg := <-client.stream.endStreamMsgs // blocking

	if msg.err != nil {
		log.Logger.Warn(
			"Stream ended due to error", "client", client, "err", msg.err,
		)
	} else {
		log.Logger.Info(
			"Stream ended gracefully", "client", client, "info", msg.info,
		)
	}

	return msg.err
}

func (net *Network) handleClientMsgs(client *client) {
	for msg := range client.stream.ackMsgs {
		client.active = true

		switch msg.Ack.(type) {

		case *pb.AckMsg_RoundLoaded:
			log.Logger.Info(
				"Client loaded round [ACK]", "client", client.username,
			)
			net.activityBitmap.SetLoadAck(client.username, true)

		case *pb.AckMsg_RoundSubmission:
			log.Logger.Info(
				"Client made a submission [ACK]", "client", client.username,
			)
			net.activityBitmap.SetSubmissionAck(client.username, true)
		}

		net.clientMsgs <- ClientMsg{client.username, msg}
	}
}

func (net *Network) BroadcastPlayerJoin(player *game.Player) {
	log.Logger.Info(
		"Broadcasting event PLAYER_JOIN", "player", player.InfoString(),
	)

	event := buildPlayerJoinedEvent(player)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastCountdown(round int, startsAt time.Time) {
	log.Logger.Info(
		"Broadcasting event COUNTING_DOWN", "round", round, "startsAt", startsAt.UTC(),
	)

	event := buildCountingDownEvent(round, startsAt)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastRoundStart(round int, endsAt time.Time) {
	log.Logger.Info(
		"Broadcasting event ROUND_STARTED", "round", round, "endsAt", endsAt.UTC(),
	)

	event := buildRoundStartedEvent(round, endsAt)
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastGameOver() {
	log.Logger.Info("Broadcasting event GAME_OVER")

	event := buildGameOverEvent()
	net.BroadcastEvent(event)
}

func (net *Network) BroadcastLoadRoundCmd(round int, challenge game.Challenge) {
	log.Logger.Info(
		"Broadcasting command LOAD_ROUND", "round", round, "challenge", challenge.InfoString(),
	)

	cmd := buildLoadRoundEvent(round, challenge)
	net.BroadcastEvent(cmd)
}

func (net *Network) BroadcastSubmitScoreCmd(round int) {
	log.Logger.Info("Broadcasting command SUBMIT_ROUND_SCORE", "round", round)

	cmd := buildSubmitRoundScoreEvent()
	net.BroadcastEvent(cmd)
}

func (net *Network) BroadcastEvent(event *pb.Event) {
	for _, client := range net.clients {
		net.sendEventToClient(event, client)
	}
}

func (net *Network) sendEventToClient(event *pb.Event, client *client) {
	if client.stream != nil {
		log.Logger.Info(
			"Sent event to client", "client", client.username,
		)
		client.stream.SendEvent(event)
	} else {
		log.Logger.Info(
			"Did not send event to client (stream is nil)", "client", client.username,
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
