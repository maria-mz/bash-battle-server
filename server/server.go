package server

import (
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
)

type Server struct {
	config config.Config

	// Registry of clients connected to the server. Identified by token.
	clients *ClientRegistry
	players *PlayerRegistry

	streamer           *Streamer
	incomingStreamMsgs <-chan IncomingMsg

	game *game.Game
}

func NewServer(config config.Config) *Server {
	incomingStreamMsgs := make(chan IncomingMsg)

	return &Server{
		config:             config,
		clients:            NewClientRegistry(),
		players:            NewPlayerRegistry(),
		streamer:           NewStreamer(incomingStreamMsgs),
		incomingStreamMsgs: incomingStreamMsgs,
		game:               game.NewGame(config.GameConfig),
	}
}

func (s *Server) Login(req *proto.LoginRequest) (*proto.LoginResponse, error) {
	log.Logger.Info("Received new login request", "username", req.Username)

	token := utils.GenerateToken()

	if err := s.game.AddNewPlayer(req.Username); err != nil {
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

func (s *Server) GetGameConfig(token string) (*proto.GameConfig, error) {
	if !s.clients.HasClient(token) {
		return &proto.GameConfig{}, errors.New("token not recognized")
	}

	return &proto.GameConfig{
		MaxPlayers:       int32(s.config.GameConfig.MaxPlayers),
		Rounds:           int32(s.config.GameConfig.Rounds),
		RoundSeconds:     int32(s.config.GameConfig.RoundDuration),
		CountdownSeconds: int32(s.config.GameConfig.CountdownDuration),
		Difficulty:       proto.Difficulty(s.config.GameConfig.Difficulty),
		FileSize:         proto.FileSize(s.config.GameConfig.FileSize),
	}, nil
}

func (s *Server) GetPlayers(token string) (*proto.Players, error) {
	if !s.clients.HasClient(token) {
		return &proto.Players{}, errors.New("token not recognized")
	}

	return &proto.Players{
		Players: s.players.GetPlayers(),
	}, nil
}

func (s *Server) Stream(token string, stream proto.BashBattle_StreamServer) error {
	if !s.clients.HasClient(token) {
		return errors.New("token not recognized")
	}

	err := s.streamer.StartStreaming(token, stream)

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
