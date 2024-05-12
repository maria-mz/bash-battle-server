package main

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	id "github.com/maria-mz/bash-battle-server/idgen"
	rg "github.com/maria-mz/bash-battle-server/registry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server represents the game server implementing the gRPC service.
type Server struct {
	pb.UnimplementedBashBattleServer

	clientRegistry *rg.ClientRegistry
	gameRegistry   *rg.GameRegistry
}

// NewServer creates a new instance of Server.
func NewServer(clientReg *rg.ClientRegistry, gameReg *rg.GameRegistry) *Server {
	return &Server{
		clientRegistry: clientReg,
		gameRegistry:   gameReg,
	}
}

func (s *Server) getUnauthenticatedErr() error {
	return status.Error(codes.Unauthenticated, "cannot identify client")
}

// getClientIDFromContext extracts the client ID from the context's metadata.
// If the token is not found, it returns an error.
func (s *Server) getClientIDFromContext(ctx context.Context) (string, error) {
	headers, _ := metadata.FromIncomingContext(ctx)
	token := headers["authorization"]

	if len(token) == 0 {
		return "", errors.New("token not found")
	}

	id := token[0]

	return id, nil
}

// Login handles the client login request.
func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	slog.Info("processing new login")

	token := id.GenerateNewToken()

	err := s.clientRegistry.RegisterClient(token, in.Name)

	if err != nil {
		switch err.(type) {

		case rg.ErrPlayerNameTaken:
			slog.Warn("login failed; name taken", "err", err)
			return &pb.LoginResponse{Status: pb.LoginStatus_NameTaken}, nil

		default:
			slog.Error("login failed", "err", err)
			return &pb.LoginResponse{}, err
		}
	}

	slog.Info("new player logged in successfully", "token", token, "name", in.Name)

	return &pb.LoginResponse{
		Status: pb.LoginStatus_LoginSuccess,
		Token:  token,
	}, nil
}

// CreateGame handles the client create game request.
func (s *Server) CreateGame(ctx context.Context, in *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	slog.Info("processing new create game request")

	clientID, err := s.getClientIDFromContext(ctx)

	if err != nil {
		slog.Warn("failed to create game; client token not found")
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	if !s.clientRegistry.HasRecord(clientID) {
		slog.Warn("failed to create game; unknown client token")
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	// TODO: build actual game plan
	plan := game.BuildTempGamePlan(int(in.GameConfig.Rounds))

	config := game.GameConfig{
		Plan:         plan,
		RoundSeconds: int(in.GameConfig.RoundMinutes), // TODO: seconds
	}

	gameID, gameCode := s.gameRegistry.RegisterGame(config)

	slog.Info("game created successfully", "id", gameID, "code", gameCode)

	return &pb.CreateGameResponse{
		GameId:   gameID,
		GameCode: gameCode,
	}, nil
}
