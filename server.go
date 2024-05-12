package main

import (
	"context"

	"github.com/maria-mz/bash-battle-proto/proto"
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
		}
	}

	return &proto.LoginResponse{
		Status: proto.LoginStatus_LoginSuccess,
		Token:  token,
	}, nil
}
