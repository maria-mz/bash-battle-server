package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/id"
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

	gameRec := GameRecord{
		GameID: id.GenerateGameID(),
		Code:   id.GenerateGameCode(),
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
