package server

import (
	"context"
	"errors"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	reg "github.com/maria-mz/bash-battle-server/registry"
	"github.com/maria-mz/bash-battle-server/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedBashBattleServer

	clients         *reg.Registry[string, ClientRecord]
	games           *reg.Registry[string, GameRecord]
	registeredNames utils.Set[string]
}

func NewServer(
	clients *reg.Registry[string, ClientRecord],
	games *reg.Registry[string, GameRecord],
) *Server {
	s := &Server{
		clients: clients,
		games:   games,
	}
	s.registeredNames = s.getRegisteredNames()

	return s
}

func (s *Server) getRegisteredNames() utils.Set[string] {
	set := utils.NewSet[string]()

	for _, record := range s.clients.Records {
		set.Add(record.PlayerName)
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
