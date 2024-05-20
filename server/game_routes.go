package server

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/id"
)

func (s *Server) CreateGame(ctx context.Context, in *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	slog.Info(fmt.Sprintf("processing create game request: %+v", in))

	_, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	// TODO: Generate actual game plan!!!
	plan := game.BuildTempGamePlan(int(in.GameConfig.Rounds))

	config := game.GameConfig{
		RoundSeconds: int(in.GameConfig.RoundSeconds),
	}

	gameID := id.GenerateGameID()
	gameCode := id.GenerateGameCode()

	gameRec := GameRecord{
		GameID: gameID,
		Code:   gameCode,
		Game:   game.NewGame(config, plan, func() {}),
	}

	s.games.WriteRecord(gameRec)

	resp := &pb.CreateGameResponse{
		GameID:   gameID,
		GameCode: gameCode,
	}

	slog.Info(fmt.Sprintf("fulfilled create game request: %+v", resp))

	return resp, nil
}

func (s *Server) JoinGame(ctx context.Context, in *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	slog.Info(fmt.Sprintf("processing join game request: %+v", in))

	clientID, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.JoinGameResponse{}, s.getUnauthenticatedErr()
	}

	gameRec, ok := s.games.GetRecord(in.GameID)

	if !ok {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_GAME_NOT_FOUND_ERR,
		}, nil
	}

	if in.GameCode != gameRec.Code {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_INVALID_CODE_ERR,
		}, nil
	}

	if gameRec.Game.State != game.InLobby {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_GAME_LOBBY_CLOSED_ERR,
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
