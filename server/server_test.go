package server

import (
	"context"
	"testing"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

const (
	testToken    = "test-token"
	testUsername = "test-player-name"
)

var testConfig = &proto.GameConfig{
	MaxPlayers:   4,
	Rounds:       10,
	RoundSeconds: 300,
	Difficulty:   proto.Difficulty_VariedDiff,
	FileSize:     proto.FileSize_VariedSize,
}

func TestMain(m *testing.M) {
	log.InitLogger()
}

func getAuthContext(token string) context.Context {
	header := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), header)
	return ctx
}

type authTest struct {
	name       string
	clients    []ClientRecord
	ctx        context.Context
	token      string
	shouldFail bool
}

func (test authTest) run(t *testing.T) {
	clients := NewRegistry[string, ClientRecord]()

	for _, client := range test.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, testConfig)

	token, err := server.authenticateClient(test.ctx)

	if test.shouldFail {
		assert.NotNil(t, err)
	} else {
		assert.Nil(t, err)
	}

	assert.Equal(t, test.token, token)
}

var authTests = []authTest{
	{
		name: "ok",
		clients: []ClientRecord{
			{
				Token:    testToken,
				Username: testUsername,
			},
		},
		ctx:        getAuthContext(testToken),
		token:      testToken,
		shouldFail: false,
	},
	{
		name:       "no token in header",
		clients:    []ClientRecord{},
		ctx:        context.Background(),
		shouldFail: true,
	},
	{
		name:       "unknown token",
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
	clients      []ClientRecord
	request      *proto.LoginRequest
	expectedResp *proto.LoginResponse
	expectedErr  error
	shouldFail   bool
}

func (test loginTest) run(t *testing.T) {
	clients := NewRegistry[string, ClientRecord]()

	for _, client := range test.clients {
		clients.WriteRecord(client)
	}

	server := NewServer(clients, testConfig)

	resp, err := server.Login(context.Background(), test.request)

	assert.Equal(t, test.expectedErr, err)
	assert.Equal(t, test.expectedResp.ErrorCode, resp.ErrorCode)

	if test.shouldFail {
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
		clients:      []ClientRecord{},
		request:      &proto.LoginRequest{Username: testUsername},
		expectedErr:  nil,
		expectedResp: &proto.LoginResponse{ErrorCode: nil},
		shouldFail:   false,
	},
	{
		name:         "name taken",
		clients:      []ClientRecord{NewClientRecord(testToken, testUsername)},
		request:      &proto.LoginRequest{Username: testUsername},
		expectedErr:  nil,
		expectedResp: &proto.LoginResponse{ErrorCode: proto.LoginResponse_ErrNameTaken.Enum()},
		shouldFail:   true,
	},
}

func TestLogin(t *testing.T) {
	for _, st := range loginTests {
		t.Run(st.name, func(t *testing.T) {
			st.run(t)
		})
	}
}
