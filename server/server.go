package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
	"google.golang.org/grpc/metadata"
)

// Server is the API for the BashBattle service.
// Implements the gRPC `BashBattleServer` interface.
type Server struct {
	proto.UnimplementedBashBattleServer

	// Registry of clients connected to the server. Identified by token.
	clients *Registry[string, ClientRecord]

	// Game instance managing the current game state
	game *game.Game

	// TODO: think about this mutex
	mutex sync.Mutex
}

func NewServer(clients *Registry[string, ClientRecord], config *proto.GameConfig) *Server {
	// TODO: make game plan random
	plan := game.BuildTempGamePlan(int(config.Rounds))

	s := &Server{
		clients: clients,
		game:    game.NewGame(config, plan, func() {}),
	}

	return s
}

func (s *Server) isNameTaken(name string) bool {
	for _, record := range s.clients.Records {
		if record.Username == name {
			return true
		}
	}

	return false
}

func (s *Server) authenticateClient(ctx context.Context) (*ClientRecord, error) {
	headers, _ := metadata.FromIncomingContext(ctx)
	auth := headers["authorization"]

	if len(auth) == 0 {
		return nil, errors.New("token not found")
	}

	token := auth[0]

	client, ok := s.clients.GetRecord(token)

	if !ok {
		return nil, errors.New("token not recognized")
	}

	return client, nil
}

func (s *Server) Login(ctx context.Context, request *proto.LoginRequest) (*proto.LoginResponse, error) {
	log.Logger.Info("New login request", "username", request.Username)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.validateLogin(request)

	if err != nil {
		log.Logger.Warn("Login failed", "reason", err)
		return &proto.LoginResponse{}, err
	}

	response := s.loginClient(request)

	return response, nil
}

func (s *Server) validateLogin(request *proto.LoginRequest) error {
	if s.isNameTaken(request.Username) {
		return ErrNameTaken{request.Username}
	}

	if s.game.State != proto.GameState_Lobby {
		return ErrGameStarted{}
	}

	if s.clients.Size() == int(s.game.Config.MaxPlayers) {
		return ErrGameFull{}
	}

	return nil
}

func (s *Server) loginClient(request *proto.LoginRequest) *proto.LoginResponse {
	token := utils.GenerateToken()

	client := NewClientRecord(token, request.Username)
	s.clients.AddRecord(*client)

	response := &proto.LoginResponse{
		Token:      token,
		Players:    s.getPlayers(),
		GameConfig: s.game.Config,
	}

	log.Logger.Info(
		"Successfully logged in client",
		"username", request.Username,
		"token", token,
	)

	s.broadcastPlayerLogin(client)

	return response
}

func (s *Server) getPlayers() []*proto.Player {
	players := make([]*proto.Player, 0, s.clients.Size())

	for _, client := range s.clients.Records {
		player := &proto.Player{
			Username: client.Username,
			Stats:    client.GameStats,
		}
		players = append(players, player)
	}

	log.Logger.Debug(fmt.Sprintf("Players = %#v", players))

	return players
}

func (s *Server) Stream(stream proto.BashBattle_StreamServer) error {
	log.Logger.Info("New call to start stream")

	client, err := s.authenticateClient(stream.Context())

	if err != nil {
		log.Logger.Warn("Failed to start stream", "err", err)
		return err
	}

	if client.Stream != nil {
		log.Logger.Warn("Stream is already running")
		return errors.New("stream is already running")
	}

	client.Stream = stream

	go s.recvStream(client)
	err = s.handleEndOfStream(client)
	return err
}

func (s *Server) recvStream(client *ClientRecord) {
	log.Logger.Info("Starting stream receive loop", "client", client.Username)

	for {
		_, err := client.Stream.Recv()

		if err == io.EOF { // happens when client calls CloseSend()
			client.EndStream <- EndStreamMsg{info: "EOF"}
			return
		}

		if err != nil {
			client.EndStream <- EndStreamMsg{err: err}
			return
		}
	}
}

func (s *Server) handleEndOfStream(client *ClientRecord) error {
	msg := <-client.EndStream // blocking

	if msg.err != nil {
		log.Logger.Warn(
			"Stream ended due to error", "client", client.Username, "err", msg.err,
		)
	} else {
		log.Logger.Info(
			"Stream ended gracefully", "client", client.Username, "info", msg.info,
		)
	}

	client.Stream = nil

	return msg.err
}

func (s *Server) broadcast(event *proto.Event) {
	log.Logger.Info("BROADCAST [start]", "event", event)

	for _, client := range s.clients.Records {
		s.sendEvent(event, client)
	}

	log.Logger.Info("BROADCAST [end]", "event", event)
}

func (s *Server) sendEvent(event *proto.Event, client *ClientRecord) {
	if client.Stream == nil {
		log.Logger.Info(
			"Skip sending event to client (no stream)", "client", client.Username,
		)
		return
	}

	err := client.Stream.Send(event)

	if err != nil {
		log.Logger.Warn(
			"Failed to send event to client", "client", client.Username,
		)
		client.EndStream <- EndStreamMsg{err: err}
	} else {
		log.Logger.Info(
			"Successfully sent event to client", "client", client.Username,
		)
	}
}

func (s *Server) broadcastPlayerLogin(client *ClientRecord) {
	player := &proto.Player{
		Username: client.Username,
		Stats:    client.GameStats,
	}

	event := &proto.Event{
		Event: &proto.Event_PlayerLogin{
			PlayerLogin: &proto.PlayerLogin{Player: player},
		},
	}

	s.broadcast(event)
}
