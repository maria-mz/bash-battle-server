package server

import (
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game/manager"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
)

type Server struct {
	config config.Config

	clients *Clients

	streamer           *Streamer
	incomingStreamMsgs <-chan IncomingMsg

	gameManager     *manager.GameManager
	gameManagerCmds <-chan manager.GameManagerCmd
}

func NewServer(config config.Config) *Server {
	incomingStreamMsgs := make(chan IncomingMsg)

	gameManager, gameManagerCmds := manager.NewGameManager(config.GameConfig)

	return &Server{
		config:             config,
		clients:            NewClients(),
		streamer:           NewStreamer(incomingStreamMsgs),
		incomingStreamMsgs: incomingStreamMsgs,
		gameManager:        gameManager,
		gameManagerCmds:    gameManagerCmds,
	}
}

func (s *Server) Connect(req *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	log.Logger.Info("New connection request", "username", req.Username)

	token := utils.GenerateToken()

	err := s.clients.AddClient(token, req.Username)

	if err != nil {
		log.Logger.Warn("Connect failed", "err", err)
		return &proto.ConnectResponse{}, err
	}

	log.Logger.Info(
		"Connected new client",
		"username", req.Username,
		"token", token,
	)

	return &proto.ConnectResponse{Token: token}, nil
}

func (s *Server) JoinGame(token string) error {
	client, ok := s.clients.GetClient(token)

	if !ok {
		return errors.New("token not recognized")
	}

	err := s.gameManager.AddPlayer(client.token, client.username)

	if err != nil {
		log.Logger.Warn(
			"Failed to add player",
			"username", client.username,
			"err", err,
		)
		return err
	}

	log.Logger.Info("Added player", "username", client.username)

	return nil
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

// TODO: fix
func (s *Server) GetPlayers(token string) (*proto.Players, error) {
	if !s.clients.HasClient(token) {
		return &proto.Players{}, errors.New("token not recognized")
	}

	return &proto.Players{
		// Players: s.players.GetPlayers(),
	}, nil
}

func (s *Server) Stream(token string, stream proto.BashBattle_StreamServer) error {
	if !s.clients.HasClient(token) {
		return errors.New("token not recognized")
	}

	err := s.streamer.StartStreaming(token, stream)

	return err
}

// TODO: fix
// func (s *Server) broadcastPlayerLogin(username string) {
// 	player, _ := s.players.GetPlayer(username)

// 	event := &proto.Event{
// 		Event: &proto.Event_PlayerLogin{
// 			PlayerLogin: &proto.PlayerLogin{Player: player},
// 		},
// 	}

// 	s.streamer.Broadcast(event, "Player Login")
// }
