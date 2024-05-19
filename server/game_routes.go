package server

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/id"
	"github.com/maria-mz/bash-battle-server/state"
	"github.com/maria-mz/bash-battle-server/utils"
)

func (s *Server) CreateGame(ctx context.Context, in *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	slog.Info(fmt.Sprintf("processing create game request: %+v", in))

	_, err := s.authenticateClient(ctx)

	if err != nil {
		slog.Warn("auth failed", "err", err)
		return &pb.CreateGameResponse{}, s.getUnauthenticatedErr()
	}

	// TODO: Generate actual game plan!!!
	plan := state.BuildTempGamePlan(int(in.GameConfig.Rounds))

	config := state.GameConfig{
		Plan:         plan,
		RoundSeconds: int(in.GameConfig.RoundSeconds),
	}

	gameID := id.GenerateGameID()
	gameCode := id.GenerateGameCode()
	store := state.NewGameStore(config)
	members := utils.NewStrSet()

	newGame := GameRecord{
		GameID:    gameID,
		GameCode:  gameCode,
		GameStore: store,
		Members:   members,
	}

	s.games.WriteRecord(newGame)

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

	client, _ := s.clients.GetRecord(clientID)

	game, ok := s.games.GetRecord(in.GameID)

	if !ok {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_GAME_NOT_FOUND_ERR,
		}, nil
	}

	if in.GameCode != game.GameCode {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_INVALID_CODE_ERR,
		}, nil
	}

	if game.GameStore.GetGameStatus() != state.InLobby {
		return &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_GAME_LOBBY_CLOSED_ERR,
		}, nil
	}

	client.GameID = &game.GameID
	game.Members.Add(client.ClientID)

	s.games.WriteRecord(game)
	s.clients.WriteRecord(client)

	return &pb.JoinGameResponse{}, nil
}
