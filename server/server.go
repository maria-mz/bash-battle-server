package server

import (
	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
)

type Server struct {
	config      config.Config
	clients     *utils.Registry[string, client]
	gameManager *gameManager
}

func NewServer(config config.Config) *Server {
	return &Server{
		config:      config,
		clients:     utils.NewRegistry[string, client](),
		gameManager: NewGameManager(config.GameConfig),
	}
}

func (s *Server) Connect(request *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	log.Logger.Info("New connect request", "request", request)

	token := utils.GenerateToken()

	nameQuery := func(client *client) bool {
		return client.username == request.Username
	}

	if s.clients.RecordsMatchingQuery(nameQuery) > 0 {
		log.Logger.Warn("Connect failed", "err", ErrUsernameTaken)
		return nil, ErrUsernameTaken
	}

	client := &client{
		token:    token,
		username: request.Username,
		active:   true,
	}

	s.clients.WriteRecord(client.token, client)

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

	if client.stream != nil {
		return ErrStreamAlreadyActive
	}

	stream := NewStream(streamSrv)
	client.stream = stream

	err := s.gameManager.ListenToClientStream(client) // Blocking

	return err
}
