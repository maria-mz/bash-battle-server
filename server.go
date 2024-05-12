package main

import (
	"context"

	"github.com/maria-mz/bash-battle-proto/proto"
	reg "github.com/maria-mz/bash-battle-server/registry"
)

// Server represents the game server implementing the gRPC service.
type Server struct {
	proto.UnimplementedBashBattleServer

	clientRegistry *reg.ClientRegistry
}

// NewServer creates a new instance of GameServer.
func NewServer(clientRegistry *reg.ClientRegistry) *Server {
	return &Server{
		clientRegistry: clientRegistry,
	}
}

// Login handles the client login request.
func (s *Server) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	token := GenerateNewToken()
	clientID := reg.ClientID(token)

	err := s.clientRegistry.RegisterClient(clientID, in.Name)

	if err != nil {
		switch err.(type) {
		case reg.ErrPlayerNameTaken:
			return &proto.LoginResponse{
				Status: proto.LoginStatus_NameTaken,
			}, nil
		default:
			return &proto.LoginResponse{}, err
		}
	}

	return &proto.LoginResponse{
		Status: proto.LoginStatus_LoginSuccess,
		Token:  token,
	}, nil
}
