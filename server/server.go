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

	clients         *reg.Registry
	games           *reg.Registry
	registeredNames utils.StrSet
}

func NewServer(clients *reg.Registry, games *reg.Registry) *Server {
	s := &Server{
		clients: clients,
		games:   games,
	}
	s.registeredNames = s.getRegisteredNames()

	return s
}

func (s *Server) getRegisteredNames() utils.StrSet {
	set := utils.NewStrSet()

	for _, rec := range s.clients.Records {
		set.Add(rec.(ClientRecord).PlayerName)
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

func (s *Server) getClientRecord(clientID string) (ClientRecord, bool) {
	record, ok := s.clients.GetRecord(clientID)
	if ok {
		return record.(ClientRecord), ok
	}
	return ClientRecord{}, ok
}

func (s *Server) getGameRecord(gameID string) (GameRecord, bool) {
	record, ok := s.games.GetRecord(gameID)
	if ok {
		return record.(GameRecord), ok
	}
	return GameRecord{}, ok
}
