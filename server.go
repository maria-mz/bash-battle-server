package main

import (
	"context"
	"errors"
	"fmt"
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

// authenticateClient extracts and validates the authorization token from the
// context headers. Returns the token if valid, otherwise returns an error.
func (s *Server) authenticateClient(ctx context.Context) (string, error) {
	headers, _ := metadata.FromIncomingContext(ctx)
	auth := headers["authorization"]

	if len(auth) == 0 {
		return "", errors.New("token not found")
	}

	token := auth[0]

	if !s.clientRegistry.HasRecord(token) {
		return "", fmt.Errorf("unknown token %s", token)
	}

	return token, nil
}

func (s *Server) getUnauthenticatedErr() error {
	return status.Error(codes.Unauthenticated, "cannot identify client")
}

// Login handles the client login request.
func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	slog.Info("processing new login")

	token := id.GenerateNewToken()

	err := s.clientRegistry.RegisterClient(token, in.Name)

	if err != nil {
		slog.Warn("login failed", "err", err)

		return &pb.LoginResponse{
			ErrorCode: pb.LoginResponse_NAME_TAKEN_ERR,
		}, nil
	}

	slog.Info("new player logged in successfully", "token", token, "name", in.Name)

	return &pb.LoginResponse{Token: token}, nil
}

// CreateGame handles the client create game request.
func (s *Server) CreateGame(ctx context.Context, in *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	slog.Info("processing create game request")

	_, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	// TODO: build actual game plan
	plan := game.BuildTempGamePlan(int(in.GameConfig.Rounds))

	config := game.GameConfig{
		Plan:         plan,
		RoundSeconds: int(in.GameConfig.RoundSeconds),
	}

	gameID, gameCode := s.gameRegistry.RegisterGame(config)

	slog.Info("game created successfully", "id", gameID, "code", gameCode)

	return &pb.CreateGameResponse{
		GameID:   gameID,
		GameCode: gameCode,
	}, nil
}

func (s *Server) JoinGame(ctx context.Context, in *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	slog.Info("processing join game request")

	clientID, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.JoinGameResponse{}, s.getUnauthenticatedErr()
	}

	playerName, _ := s.clientRegistry.GetPlayerName(clientID)

	err = s.gameRegistry.JoinGame(in.GameID, in.GameCode, playerName)

	switch e := err.(type) {

	case rg.ErrGameNotFound:
		slog.Warn("game not found", "err", e)

		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_GAME_NOT_FOUND_ERR,
		}, nil

	case rg.ErrInvalidCode:
		slog.Warn("invalid code", "err", e)

		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_INVALID_CODE_ERR,
		}, nil

	case rg.ErrJoinAfterLobbyClosed:
		slog.Warn("lobby closed", "err", e)

		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_GAME_LOBBY_CLOSED_ERR,
		}, nil
	}

	return &pb.JoinGameResponse{}, nil
}
