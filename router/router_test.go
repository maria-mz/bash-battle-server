package router

import (
	"context"
	"testing"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
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

func getAuthContext(token string) context.Context {
	header := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), header)
	return ctx
}

type authTest struct {
	name       string
	ctx        context.Context
	token      string
	shouldFail bool
}

func (test authTest) run(t *testing.T) {
	server := server.NewServer(testConfig)
	router := NewServerRouter(server)

	token := router.getClientToken(test.ctx)

	if test.shouldFail {
		assert.Equal(t, NoToken, token)
	} else {
		assert.Equal(t, test.token, token)
	}
}

var authTests = []authTest{
	{
		name:       "token in header",
		ctx:        getAuthContext("test-token"),
		token:      "test-token",
		shouldFail: false,
	},
	{
		name:       "token not in header",
		ctx:        context.Background(),
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
