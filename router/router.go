package router

import (
	"context"
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/server"
	"google.golang.org/grpc/metadata"
)

// ServerRouter is the API for the BashBattle gRPC service.
// Handles authorization, directs processing to the internal server.
type ServerRouter struct {
	proto.UnimplementedBashBattleServer
	server *server.Server
}

func NewServerRouter(s *server.Server) *ServerRouter {
	return &ServerRouter{server: s}
}

func (s *ServerRouter) getToken(ctx context.Context) (string, bool) {
	var token string

	headers, _ := metadata.FromIncomingContext(ctx)
	auth := headers["authorization"]

	if len(auth) == 0 {
		return token, false
	}

	token = auth[0]

	return token, true
}

func (s *ServerRouter) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	res, err := s.server.Login(req)
	return res, err
}

func (s *ServerRouter) Stream(stream proto.BashBattle_StreamServer) error {
	token, ok := s.getToken(stream.Context())

	if !ok {
		return errors.New("token not found")
	}

	err := s.server.Stream(token, stream)

	return err
}
