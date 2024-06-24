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
		Rounds:            2,
		RoundDuration:     2,
		CountdownDuration: 1,
	},
}

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

type connectTest struct {
	name       string
	requests   []*proto.ConnectRequest
	shouldFail bool
}

func (test connectTest) run(t *testing.T) {
	server := NewServer(testConfig)

	for i := 0; i < len(test.requests)-1; i++ {
		server.Connect(test.requests[i])
	}

	requestToTest := test.requests[len(test.requests)-1]

	resp, err := server.Connect(requestToTest)

	if test.shouldFail {
		assert.Nil(t, resp)
		assert.NotNil(t, err)
	} else {
		assert.NotNil(t, resp)
		assert.NotEqual(t, resp.Token, "")
		assert.Nil(t, err)
		assert.True(t, server.clients.HasRecord(resp.Token))
	}
}

var connectTests = []connectTest{
	{
		name: "first connect",
		requests: []*proto.ConnectRequest{
			{Username: "player-1"},
		},
		shouldFail: false,
	},
	{
		name: "all players connect",
		requests: []*proto.ConnectRequest{
			{Username: "player-1"},
			{Username: "player-2"},
			{Username: "player-3"},
		},
		shouldFail: false,
	},
	{
		name: "name taken",
		requests: []*proto.ConnectRequest{
			{Username: "player-1"},
			{Username: "player-1"},
		},
		shouldFail: true,
	},
	{
		name: "more than max players",
		requests: []*proto.ConnectRequest{
			{Username: "player-1"},
			{Username: "player-2"},
			{Username: "player-3"},
			{Username: "player-4"},
		},
		shouldFail: false,
	},
}

func TestConnect(t *testing.T) {
	for _, st := range connectTests {
		t.Run(st.name, func(t *testing.T) {
			st.run(t)
		})
	}
}
