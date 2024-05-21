package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server is the API for the BashBattle service.
// Implements the gRPC `BashBattleServer` interface.
type Server struct {
	proto.UnimplementedBashBattleServer

	// Registry of clients connected to the server. Identified by token.
	clients *Registry[string, ClientRecord]

	// Set of usernames currently in use
	usedNames utils.Set[string]

	// Game instance managing the current game state
	game *game.Game

	// TODO: think about this mutex
	mutex sync.Mutex
}

func NewServer(clients *Registry[string, ClientRecord], config game.GameConfig) *Server {
	// TODO: make game plan random
	plan := game.BuildTempGamePlan(int(config.Rounds))

	s := &Server{
		clients: clients,
		game:    game.NewGame(config, plan, func() {}),
	}
	s.usedNames = s.getUsedNames()
	return s
}

func (s *Server) getUsedNames() utils.Set[string] {
	set := utils.NewSet[string]()

	for _, record := range s.clients.Records {
		set.Add(record.Username)
	}

	return set
}

func (s *Server) authenticateClient(ctx context.Context) (string, error) {
	headers, _ := metadata.FromIncomingContext(ctx)
	auth := headers["authorization"]

	if len(auth) == 0 {
		return "", errors.New("token not found")
	}

	token := auth[0]

	if !s.clients.HasRecord(token) {
		return "", errors.New("unknown token")
	}

	return token, nil
}

func (s *Server) getUnauthenticatedErr() error {
	return status.Error(codes.Unauthenticated, "unauthorized")
}

func (s *Server) Login(ctx context.Context, request *proto.LoginRequest) (*proto.LoginResponse, error) {
	log.Logger.Info("New login request", "username", request.Username)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.validateLogin(request)

	if err != nil {
		log.Logger.Warn("Login failed", "reason", err)
		return &proto.LoginResponse{ErrorCode: err}, nil
	}

	response := s.loginClient(request)

	return response, nil
}

func (s *Server) validateLogin(request *proto.LoginRequest) *proto.LoginResponse_ErrorCode {
	if s.usedNames.Contains(request.Username) {
		return proto.LoginResponse_ErrNameTaken.Enum()
	}

	if s.game.State != game.InLobby {
		return proto.LoginResponse_ErrGameStarted.Enum()
	}

	if s.clients.Size() == s.game.Config.MaxPlayers {
		return proto.LoginResponse_ErrGameFull.Enum()
	}

	return nil
}

func (s *Server) loginClient(request *proto.LoginRequest) *proto.LoginResponse {
	token := GenerateNewToken()

	client := ClientRecord{
		Token:     token,
		Username:  request.Username,
		GameStats: game.NewGameStats(),
	}

	s.clients.WriteRecord(client)
	s.usedNames.Add(client.Username)

	response := &proto.LoginResponse{Token: token, Players: s.getPlayers()}

	log.Logger.Info(
		"Successfully logged in client",
		"username", request.Username,
		"token", token,
	)

	return response
}

func (s *Server) getPlayers() []*proto.Player {
	players := make([]*proto.Player, 0, s.clients.Size())

	for _, record := range s.clients.Records {
		player := &proto.Player{
			Username: record.Username,
			Stats:    record.GameStats,
		}
		players = append(players, player)
	}

	log.Logger.Debug(fmt.Sprintf("Players = %#v", players))

	return players
}
