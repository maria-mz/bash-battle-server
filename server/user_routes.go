package server

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/id"
)

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	slog.Info(fmt.Sprintf("processing login request: %+v", req))

	if s.registeredNames.Contains(req.PlayerName) {
		return &pb.LoginResponse{
			ErrorCode: pb.LoginResponse_ErrNameTaken.Enum(),
		}, nil
	}

	token := id.GenerateNewToken()

	newClient := ClientRecord{
		ClientID:   token,
		PlayerName: req.PlayerName,
	}

	s.clients.WriteRecord(newClient)
	s.registeredNames.Add(req.PlayerName)

	resp := &pb.LoginResponse{Token: token}
	slog.Info(fmt.Sprintf("fulfilled login request: %+v", resp))

	return resp, nil
}
