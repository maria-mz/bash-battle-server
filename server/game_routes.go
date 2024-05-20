package server

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/id"
)

func (s *Server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	slog.Info(fmt.Sprintf("processing create game request: %+v", req))

	_, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	config := game.GameConfig{
		Rounds:       int(req.GameConfig.Rounds),
		RoundSeconds: int(req.GameConfig.RoundSeconds),
	}

	gameRec := s.createGame(config)
	s.games.WriteRecord(gameRec)

	resp := &pb.CreateGameResponse{
		GameID:   gameRec.GameID,
		GameCode: gameRec.Code,
	}

	slog.Info(fmt.Sprintf("fulfilled create game request: %+v", resp))

	return resp, nil
}

func (s *Server) createGame(config game.GameConfig) GameRecord {
	// TODO: Generate actual game plan!!!
	plan := game.BuildTempGamePlan(config.Rounds)

	gameID := id.GenerateGameID()
	gameCode := id.GenerateGameCode()

	gameRec := GameRecord{
		GameID: gameID,
		Code:   gameCode,
		Game:   game.NewGame(config, plan, func() {}),
	}

	return gameRec
}

func (s *Server) JoinGame(ctx context.Context, req *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	slog.Info(fmt.Sprintf("processing join game request: %+v", req))

	clientID, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.JoinGameResponse{}, s.getUnauthenticatedErr()
	}

	gameRec, ok := s.games.GetRecord(req.GameID)

	if !ok {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_ErrGameNotFound.Enum(),
		}, nil
	}

	if req.GameCode != gameRec.Code {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_ErrInvalidCode.Enum(),
		}, nil
	}

	if gameRec.Game.State != game.InLobby {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_ErrJoinsClosed.Enum(),
		}, nil
	}

	clientRec, _ := s.clients.GetRecord(clientID)
	s.addClientToGame(clientRec, gameRec)

	return &pb.JoinGameResponse{}, nil
}

func (s *Server) addClientToGame(clientRec ClientRecord, gameRec GameRecord) {
	clientRec.GameID = &gameRec.GameID

	player := game.NewPlayer(clientRec.ClientID, clientRec.PlayerName)
	gameRec.Game.Players.WriteRecord(player)

	s.games.WriteRecord(gameRec)
	s.clients.WriteRecord(clientRec)
}
