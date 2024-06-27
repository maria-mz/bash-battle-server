package server

import (
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server/game_manager"
	"github.com/maria-mz/bash-battle-server/server/network"
	"github.com/maria-mz/bash-battle-server/utils"
)

var ErrTokenNotRecognized = errors.New("token not recognized")
var ErrStreamAlreadyActive = errors.New("stream is already active")
var ErrUsernameTaken = errors.New("a player with this name already exists")

type Server struct {
	config       config.Config
	clients      map[string]*network.Client
	usernamePool utils.Set[string]
	gameManager  *game_manager.GameManager
}

func NewServer(config config.Config) *Server {
	return &Server{
		config:       config,
		clients:      make(map[string]*network.Client),
		usernamePool: utils.NewSet[string](),
		gameManager:  game_manager.NewGameManager(config.GameConfig),
	}
}

func (s *Server) Connect(request *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	log.Logger.Info("New connect request", "request", request)

	if s.usernamePool.Contains(request.Username) {
		log.Logger.Warn("Connect failed", "err", ErrUsernameTaken)
		return nil, ErrUsernameTaken
	}

	token := utils.GenerateToken()

	client := &network.Client{
		Token:    token,
		Username: request.Username,
		Active:   true,
	}

	s.clients[client.Token] = client
	s.usernamePool.Add(request.Username)

	log.Logger.Info("Connected new client", "client", client)

	return &proto.ConnectResponse{Token: token}, nil
}

func (s *Server) JoinGame(token string) error {
	log.Logger.Info("New join game request")

	client, ok := s.clients[token]

	if !ok {
		log.Logger.Info("Failed to join game", "err", ErrTokenNotRecognized)
		return ErrTokenNotRecognized
	}

	err := s.gameManager.AddClient(client)

	if err != nil {
		log.Logger.Warn("Failed to join game", "client", client, "err", err)
		return err
	}

	log.Logger.Info("Client joined game", "client", client)

	return nil
}

func (s *Server) GetGameConfig(token string) (*proto.GameConfig, error) {
	_, ok := s.clients[token]
	if !ok {
		return nil, ErrTokenNotRecognized
	}

	return &proto.GameConfig{
		MaxPlayers:   int32(s.config.GameConfig.MaxPlayers),
		Rounds:       int32(s.config.GameConfig.Rounds),
		RoundSeconds: int32(s.config.GameConfig.RoundDuration),
		Difficulty:   proto.Difficulty(s.config.GameConfig.Difficulty),
		FileSize:     proto.FileSize(s.config.GameConfig.FileSize),
	}, nil
}

func (s *Server) Stream(token string, streamSrv proto.BashBattle_StreamServer) error {
	client, ok := s.clients[token]
	if !ok {
		return ErrTokenNotRecognized
	}

	if client.Stream != nil {
		return ErrStreamAlreadyActive
	}

	stream := network.NewStream(streamSrv)
	client.Stream = stream

	err := s.gameManager.ListenForClientMsgs(client.Username) // Blocking

	return err
}
