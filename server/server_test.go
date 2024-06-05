package server

import (
	"testing"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/stretchr/testify/assert"
)

var testConfig = config.Config{
	GameConfig: config.GameConfig{
		MaxPlayers:        3,
		Rounds:            5,
		RoundDuration:     300,
		CountdownDuration: 10,
	},
}

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

type loginTest struct {
	name       string
	requests   []*proto.LoginRequest
	shouldFail bool
}

func (test loginTest) run(t *testing.T) {
	server := NewServer(testConfig)

	for i := 0; i < len(test.requests)-1; i++ {
		server.Login(test.requests[i])
	}

	requestToTest := test.requests[len(test.requests)-1]

	resp, err := server.Login(requestToTest)

	if test.shouldFail {
		assert.Equal(t, resp.Token, "")
		assert.NotNil(t, err)
	} else {
		assert.NotEqual(t, resp.Token, "")
		assert.Nil(t, err)
		assert.True(t, server.clients.HasClient(resp.Token))
		assert.True(t, server.players.HasPlayer(requestToTest.Username))
	}
}

var loginTests = []loginTest{
	{
		name: "first login",
		requests: []*proto.LoginRequest{
			{Username: "player-1"},
		},
		shouldFail: false,
	},
	{
		name: "all players login",
		requests: []*proto.LoginRequest{
			{Username: "player-1"},
			{Username: "player-2"},
			{Username: "player-3"},
		},
		shouldFail: false,
	},
	{
		name: "name taken",
		requests: []*proto.LoginRequest{
			{Username: "player-1"},
			{Username: "player-1"},
		},
		shouldFail: true,
	},
	{
		name: "too many players",
		requests: []*proto.LoginRequest{
			{Username: "player-1"},
			{Username: "player-2"},
			{Username: "player-3"},
			{Username: "player-4"},
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

// TODO: Add stream tests
