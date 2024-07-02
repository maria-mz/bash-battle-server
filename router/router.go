package router

import (
	"context"
	"errors"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/server"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrTokenNotFound = errors.New("token is missing")

// ServerRouter is the API for the BashBattle gRPC service.
// Handles authorization, directs processing of calls to the internal server.
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

func (s *ServerRouter) Connect(ctx context.Context, in *proto.ConnectRequest) (*proto.ConnectResponse, error) {
	res, err := s.server.Connect(in)
	return res, err
}

func (s *ServerRouter) JoinGame(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	token, ok := s.getToken(ctx)

	if !ok {
		return &emptypb.Empty{}, ErrTokenNotFound
	}

	err := s.server.JoinGame(token)

	return &emptypb.Empty{}, err
}

func (s *ServerRouter) GetGameConfig(ctx context.Context, _ *emptypb.Empty) (*proto.GameConfig, error) {
	token, ok := s.getToken(ctx)

	if !ok {
		return &proto.GameConfig{}, ErrTokenNotFound
	}

	config, err := s.server.GetGameConfig(token)

	return config, err
}

func (s *ServerRouter) GetPlayers(ctx context.Context, _ *emptypb.Empty) (*proto.Players, error) {
	token, ok := s.getToken(ctx)

	if !ok {
		return &proto.Players{}, ErrTokenNotFound
	}

	players, err := s.server.GetPlayers(token)

	return players, err
}

func (s *ServerRouter) Stream(stream proto.BashBattle_StreamServer) error {
	token, ok := s.getToken(stream.Context())

	if !ok {
		return ErrTokenNotFound
	}

	err := s.server.Stream(token, stream)

	return err
}
