package server

import (
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game/blueprint"
	"github.com/maria-mz/bash-battle-server/game/fsm"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
)

type Server struct {
	config config.Config

	// Registry of clients connected to the server. Identified by token.
	clients *ClientRegistry
	players *PlayerRegistry

	streamer   *Streamer
	streamMsgs <-chan StreamMsg

	blueprint blueprint.Blueprint

	game       *fsm.FSM
	fsmUpdates <-chan fsm.FSMState
}

func NewServer(config config.Config) *Server {
	updates := make(chan fsm.FSMState)
	streamMsgs := make(chan StreamMsg)

	return &Server{
		config:     config,
		clients:    NewClientRegistry(),
		players:    NewPlayerRegistry(),
		streamer:   NewStreamer(streamMsgs),
		streamMsgs: streamMsgs,
		game:       fsm.NewFSM(config.GameConfig, updates),
		blueprint:  blueprint.BuildBlueprint(config.GameConfig),
		fsmUpdates: updates,
	}
}

func (s *Server) Login(req *proto.LoginRequest) (*proto.LoginResponse, error) {
	log.Logger.Info("Received new login request", "username", req.Username)

	token := utils.GenerateToken()

	if err := s.game.AddPlayer(req.Username); err != nil {
		log.Logger.Warn("Login failed", "err", err)
		return &proto.LoginResponse{}, err
	}

	s.clients.AddClient(token, req.Username)
	s.players.AddPlayer(req.Username)

	s.broadcastPlayerLogin(req.Username)

	log.Logger.Info(
		"Successfully logged in client",
		"username", req.Username,
		"token", token,
	)

	return &proto.LoginResponse{Token: token}, nil
}

func (s *Server) Stream(token string, stream proto.BashBattle_StreamServer) error {
	_, ok := s.clients.GetClient(token)

	if !ok {
		return errors.New("client not recognized")
	}

	if s.streamer.IsStreamActive(token) {
		return errors.New("stream is already running")
	}

	s.streamer.UnRegisterStream(token) // Make sure old stream is gone
	s.streamer.RegisterStream(token, stream)

	err := s.streamer.StartStreaming(token)

	return err
}

func (s *Server) broadcastPlayerLogin(username string) {
	player, _ := s.players.GetPlayer(username)

	event := &proto.Event{
		Event: &proto.Event_PlayerLogin{
			PlayerLogin: &proto.PlayerLogin{Player: player},
		},
	}

	s.streamer.Broadcast(event, "Player Login")
}
