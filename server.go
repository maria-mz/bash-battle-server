package main

import (
	"context"

	"github.com/maria-mz/bash-battle-proto/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GameServer represents the game server implementing the gRPC service.
type GameServer struct {
	proto.UnimplementedBashBattleServer

	clientRegistry *ClientRegistry
}

// NewGameServer creates a new instance of GameServer.
func NewGameServer(clientRegistry *ClientRegistry) *GameServer {
	return &GameServer{
		clientRegistry: clientRegistry,
	}
}

// getInternalErr returns an internal server error as a gRPC status error.
// Note, use for serious errors.
func (s *GameServer) getInternalErr() error {
	return status.Error(codes.Internal, "internal server error")
}

// Login handles the client login request.
func (s *GameServer) Login(ctx context.Context, request *proto.LoginRequest) (*proto.LoginResponse, error) {
	token := GenerateNewToken()
	clientID := ClientID(token)

	err := s.clientRegistry.RegisterClient(clientID, request.Name)

	if err != nil {
		switch err.(type) {
		case ErrPlayerNameTaken:
			return &proto.LoginResponse{
				Status: proto.LoginStatus_NameTaken,
			}, nil
		default:
			return nil, s.getInternalErr()
		}
	}

	return &proto.LoginResponse{
		Status: proto.LoginStatus_LoginSuccess,
		Token:  token,
	}, nil
}
