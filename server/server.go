package server

import (
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server/client"
	"github.com/maria-mz/bash-battle-server/server/game_manager"
	"github.com/maria-mz/bash-battle-server/server/stream"
	"github.com/maria-mz/bash-battle-server/utils"
)

var ErrTokenNotRecognized = errors.New("token not recognized")
var ErrStreamAlreadyActive = errors.New("stream is already active")
var ErrUsernameTaken = errors.New("a player with this name already exists")

type Server struct {
	config      config.Config
	clients     *utils.Registry[string, client.Client]
	gameManager *game_manager.GameManager
}

func NewServer(config config.Config) *Server {
	return &Server{
		config:      config,
		clients:     utils.NewRegistry[string, client.Client](),
		gameManager: game_manager.NewGameManager(config.GameConfig),
	}
}

func (s *Server) Connect(request *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	log.Logger.Info("New connect request", "request", request)

	token := utils.GenerateToken()

	nameQuery := func(client *client.Client) bool {
		return client.Username == request.Username
	}

	if s.clients.RecordsMatchingQuery(nameQuery) > 0 {
		log.Logger.Warn("Connect failed", "err", ErrUsernameTaken)
		return nil, ErrUsernameTaken
	}

	client := &client.Client{
		Token:    token,
		Username: request.Username,
		Active:   true,
	}

	s.clients.WriteRecord(client.Token, client)

	log.Logger.Info("Connected new client", "client", client)

	return &proto.ConnectResponse{Token: token}, nil
}

func (s *Server) JoinGame(token string) error {
	log.Logger.Info("New join game request")

	client, ok := s.clients.GetRecord(token)

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
	if !s.clients.HasRecord(token) {
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
	client, ok := s.clients.GetRecord(token)

	if !ok {
		return ErrTokenNotRecognized
	}

	if client.Stream != nil {
		return ErrStreamAlreadyActive
	}

	stream := stream.NewStream(streamSrv)
	client.Stream = stream

	err := s.gameManager.ListenForClientMsgs(client.Username) // Blocking

	return err
}
