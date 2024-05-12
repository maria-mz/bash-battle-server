package main

import (
	"context"
	"errors"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	idgen "github.com/maria-mz/bash-battle-server/idgen"
	reg "github.com/maria-mz/bash-battle-server/registry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server represents the game server implementing the gRPC service.
type Server struct {
	pb.UnimplementedBashBattleServer

	clientRegistry *reg.ClientRegistry
	gameRegistry   *reg.GameRegistry
}

// NewServer creates a new instance of Server.
func NewServer(clientReg *reg.ClientRegistry, gameReg *reg.GameRegistry) *Server {
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
	token := idgen.GenerateNewToken()

	err := s.clientRegistry.RegisterClient(token, in.Name)

	if err != nil {
		switch err.(type) {
		case reg.ErrPlayerNameTaken:
			return &pb.LoginResponse{
				Status: pb.LoginStatus_NameTaken,
			}, nil
		default:
			return &pb.LoginResponse{}, err
		}
	}

	return &pb.LoginResponse{
		Status: pb.LoginStatus_LoginSuccess,
		Token:  token,
	}, nil
}

// CreateGame handles the client create game request.
func (s *Server) CreateGame(ctx context.Context, in *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	clientID, err := s.getClientIDFromContext(ctx)

	if err != nil || !s.clientRegistry.HasRecord(clientID) {
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	plan := game.BuildTempGamePlan(int(in.GameConfig.Rounds))

	config := game.GameConfig{
		Plan:         plan,
		RoundSeconds: int(in.GameConfig.RoundMinutes),
	}

	gameID, gameCode := s.gameRegistry.RegisterGame(config)

	return &pb.CreateGameResponse{
		GameId:   string(gameID),
		GameCode: gameCode,
	}, nil
}
