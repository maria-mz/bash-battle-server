package server

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/id"
)

func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	slog.Info(fmt.Sprintf("processing login request: %+v", in))

	if s.registeredNames.Contains(in.PlayerName) {
		return &pb.LoginResponse{
			ErrorCode: pb.LoginResponse_NAME_TAKEN_ERR,
		}, nil
	}

	token := id.GenerateNewToken()

	newClient := ClientRecord{
		ClientID:   token,
		PlayerName: in.PlayerName,
	}

	s.clients.WriteRecord(newClient)
	s.registeredNames.Add(in.PlayerName)

	resp := &pb.LoginResponse{Token: token}
	slog.Info(fmt.Sprintf("fulfilled login request: %+v", resp))

	return resp, nil
}
