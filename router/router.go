package router

import (
	"context"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/server"
	"google.golang.org/grpc/metadata"
)

const NoToken = ""

// ServerRouter is the API for the BashBattle gRPC service.
// Handles authorization, directs processing to the internal server.
type ServerRouter struct {
	proto.UnimplementedBashBattleServer
	server *server.Server
}

func NewServerRouter(s *server.Server) *ServerRouter {
	return &ServerRouter{server: s}
}

func (s *ServerRouter) getClientToken(ctx context.Context) string {
	headers, _ := metadata.FromIncomingContext(ctx)
	auth := headers["authorization"]

	if len(auth) == 0 {
		return NoToken
	}

	token := auth[0]

	return token
}

func (s *ServerRouter) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	res, err := s.server.Login(req)
	return res, err
}

func (s *ServerRouter) Stream(stream proto.BashBattle_StreamServer) error {
	token := s.getClientToken(stream.Context())

	err := s.server.Stream(token, stream)

	return err
}
