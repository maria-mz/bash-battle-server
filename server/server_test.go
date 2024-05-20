package server

import (
	"context"
	"testing"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	reg "github.com/maria-mz/bash-battle-server/registry"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

const (
	testToken      = "test-token"
	testPlayerName = "test-player-name"
	testGameID     = "test-game-id"
	testGameCode   = "test-game-code"
)

func getAuthContext(token string) context.Context {
	header := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), header)
	return ctx
}

type authTest struct {
	name       string
	games      []GameRecord
	clients    []ClientRecord
	ctx        context.Context
	token      string
	shouldFail bool
}

func (st authTest) run(t *testing.T) {
	games := reg.NewRegistry[string, GameRecord]()
	clients := reg.NewRegistry[string, ClientRecord]()

	for _, game := range st.games {
		games.WriteRecord(game)
	}

	for _, client := range st.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, games)

	token, err := server.authenticateClient(st.ctx)

	if st.shouldFail {
		assert.NotNil(t, err)
	} else {
		assert.Nil(t, err)
	}

	assert.Equal(t, st.token, token)
}

var authTests = []authTest{
	{
		name:  "ok",
		games: []GameRecord{},
		clients: []ClientRecord{
			{
				ClientID:   testToken,
				PlayerName: testPlayerName,
			},
		},
		ctx:        getAuthContext(testToken),
		token:      testToken,
		shouldFail: false,
	},
	{
		name:       "no token in header",
		games:      []GameRecord{},
		clients:    []ClientRecord{},
		ctx:        context.Background(),
		shouldFail: true,
	},
	{
		name:       "unknown token",
		games:      []GameRecord{},
		clients:    []ClientRecord{},
		ctx:        getAuthContext(testToken),
		shouldFail: true,
	},
}

func TestAuth(t *testing.T) {
	for _, st := range authTests {
		t.Run(st.name, func(t *testing.T) {
			st.run(t)
		})
	}
}

type loginTest struct {
	name         string
	games        []GameRecord
	clients      []ClientRecord
	request      *pb.LoginRequest
	expectedResp *pb.LoginResponse
	expectedErr  error
	shouldFail   bool
}

func (st loginTest) run(t *testing.T) {
	games := reg.NewRegistry[string, GameRecord]()
	clients := reg.NewRegistry[string, ClientRecord]()

	for _, game := range st.games {
		games.WriteRecord(game)
	}

	for _, client := range st.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, games)

	resp, err := server.Login(context.Background(), st.request)

	assert.Equal(t, st.expectedErr, err)
	assert.Equal(t, st.expectedResp.ErrorCode, resp.ErrorCode)

	if st.shouldFail {
		assert.Equal(t, resp.Token, "")
		assert.False(t, clients.HasRecord(resp.Token))
	} else {
		assert.NotEqual(t, resp.Token, "")
		assert.True(t, clients.HasRecord(resp.Token))
	}
}

var loginTests = []loginTest{
	{
		name:         "first client + ok",
		games:        []GameRecord{},
		clients:      []ClientRecord{},
		request:      &pb.LoginRequest{PlayerName: testPlayerName},
		expectedErr:  nil,
		expectedResp: &pb.LoginResponse{ErrorCode: nil},
		shouldFail:   false,
	},
	{
		name:  "name taken",
		games: []GameRecord{},
		clients: []ClientRecord{
			{
				ClientID:   testToken,
				PlayerName: testPlayerName,
			},
		},
		request:     &pb.LoginRequest{PlayerName: testPlayerName},
		expectedErr: nil,
		expectedResp: &pb.LoginResponse{
			ErrorCode: pb.LoginResponse_ErrNameTaken.Enum(),
			// Note: would add token... but it's random so can't really check it
		},
		shouldFail: true,
	},
}

func TestLogin(t *testing.T) {
	for _, st := range loginTests {
		t.Run(st.name, func(t *testing.T) {
			st.run(t)
		})
	}
}

type createGameTest struct {
	name        string
	games       []GameRecord
	clients     []ClientRecord
	ctx         context.Context
	request     *pb.CreateGameRequest
	expectedErr error
}

func (st createGameTest) run(t *testing.T) {
	games := reg.NewRegistry[string, GameRecord]()
	clients := reg.NewRegistry[string, ClientRecord]()

	for _, game := range st.games {
		games.WriteRecord(game)
	}

	for _, client := range st.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, games)

	resp, err := server.CreateGame(st.ctx, st.request)

	assert.Equal(t, st.expectedErr, err)
	assert.NotEqual(t, resp.GameCode, "")
	assert.True(t, games.HasRecord(resp.GameID))
}

var createGameTests = []createGameTest{
	{
		name:  "first game + ok",
		games: []GameRecord{},
		clients: []ClientRecord{
			{
				ClientID:   testToken,
				PlayerName: testPlayerName,
			},
		},
		ctx: getAuthContext(testToken),
		request: &pb.CreateGameRequest{
			GameConfig: &pb.GameConfig{
				MaxPlayers:   5,
				Rounds:       10,
				RoundSeconds: 300,
			},
		},
		expectedErr: nil,
	},
}

func TestCreateGame(t *testing.T) {
	for _, st := range createGameTests {
		t.Run(st.name, func(t *testing.T) {
			st.run(t)
		})
	}
}

type joinGameTest struct {
	name         string
	games        []GameRecord
	clients      []ClientRecord
	ctx          context.Context
	request      *pb.JoinGameRequest
	expectedResp *pb.JoinGameResponse
	expectedErr  error
	shouldFail   bool
}

func (st joinGameTest) run(t *testing.T) {
	games := reg.NewRegistry[string, GameRecord]()
	clients := reg.NewRegistry[string, ClientRecord]()

	for _, game := range st.games {
		games.WriteRecord(game)
	}

	for _, client := range st.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, games)

	resp, err := server.JoinGame(st.ctx, st.request)

	assert.Equal(t, st.expectedErr, err)
	assert.Equal(t, st.expectedResp.ErrorCode, resp.ErrorCode)

	// TODO: check if in members
}

var joinGameTests = []joinGameTest{
	{
		name: "ok",
		games: []GameRecord{
			{
				GameID: testGameID,
				Code:   testGameCode,
				Game:   game.NewGame(game.GameConfig{}, game.GamePlan{}, func() {}),
			},
		},
		clients: []ClientRecord{
			{
				ClientID:   testToken,
				PlayerName: testPlayerName,
			},
		},
		ctx: getAuthContext(testToken),
		request: &pb.JoinGameRequest{
			GameID:   testGameID,
			GameCode: testGameCode,
		},
		expectedResp: &pb.JoinGameResponse{ErrorCode: nil},
		expectedErr:  nil,
		shouldFail:   false,
	},
	{
		name: "invalid code",
		games: []GameRecord{
			{
				GameID: testGameID,
				Code:   testGameCode,
				Game:   game.NewGame(game.GameConfig{}, game.GamePlan{}, func() {}),
			},
		},
		clients: []ClientRecord{
			{
				ClientID:   testToken,
				PlayerName: testPlayerName,
			},
		},
		ctx: getAuthContext(testToken),
		request: &pb.JoinGameRequest{
			GameID:   testGameID,
			GameCode: "some-invalid-code",
		},
		expectedResp: &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_ErrInvalidCode.Enum(),
		},
		expectedErr: nil,
		shouldFail:  true,
	},
	{
		name: "game not found",
		games: []GameRecord{
			{
				GameID: testGameID,
				Code:   testGameCode,
				Game:   game.NewGame(game.GameConfig{}, game.GamePlan{}, func() {}),
			},
		},
		clients: []ClientRecord{
			{
				ClientID:   testToken,
				PlayerName: testPlayerName,
			},
		},
		ctx: getAuthContext(testToken),
		request: &pb.JoinGameRequest{
			GameID:   "some-unknown-game-id",
			GameCode: testGameCode,
		},
		expectedResp: &pb.JoinGameResponse{
			ErrorCode: pb.JoinGameResponse_ErrGameNotFound.Enum(),
		},
		expectedErr: nil,
		shouldFail:  true,
	},
}

func TestJoinGame(t *testing.T) {
	for _, st := range joinGameTests {
		t.Run(st.name, func(t *testing.T) {
			st.run(t)
		})
	}
}
