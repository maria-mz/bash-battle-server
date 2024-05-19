package server

import (
	"context"
	"testing"

	reg "github.com/maria-mz/bash-battle-server/registry"
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
	games := reg.NewRegistry()
	clients := reg.NewRegistry()

	for _, game := range st.games {
		games.WriteRecord(game)
	}

	for _, client := range st.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, games)

	token, err := server.authenticateClient(st.ctx)

	if !st.shouldFail && err != nil {
		t.Errorf("expected no error but got %s", err)
	}

	if st.shouldFail && err == nil {
		t.Errorf("expected error but got no error")
	}

	if token != st.token {
		t.Errorf("token mismatch: actual != expected: %s != %s", token, st.token)
	}
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
