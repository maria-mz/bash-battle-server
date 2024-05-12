package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/maria-mz/bash-battle-proto/proto"
	reg "github.com/maria-mz/bash-battle-server/registry"
	"google.golang.org/grpc/metadata"
)

var TEST_TOKEN = "test-token"
var TEST_PLAYER_NAME = "test-player-name"

func getAuthContext(token string) context.Context {
	header := metadata.New(map[string]string{"authorization": token})
	ctx := metadata.NewIncomingContext(context.Background(), header)
	return ctx
}

func TestAuthenticate_Success(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()
	server := NewServer(clientRegistry, gameRegistry)

	clientRegistry.RegisterClient(TEST_TOKEN, TEST_PLAYER_NAME)

	ctx := getAuthContext(TEST_TOKEN)

	token, err := server.authenticateClient(ctx)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if token != TEST_TOKEN {
		t.Fatalf("%s != %s", token, TEST_TOKEN)
	}
}

func TestAuthenticate_NoToken(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()
	server := NewServer(clientRegistry, gameRegistry)

	_, err := server.authenticateClient(context.Background())

	expectedErr := "no token found"
	if err.Error() == expectedErr {
		t.Fatalf("%s != %s", err, expectedErr)
	}
}

func TestAuthenticate_NoRecord(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()
	server := NewServer(clientRegistry, gameRegistry)

	ctx := getAuthContext(TEST_TOKEN)

	_, err := server.authenticateClient(ctx)

	expectedErr := fmt.Sprintf("unknown token %s", TEST_TOKEN)
	if err.Error() != expectedErr {
		t.Fatalf("%s != %s", err, expectedErr)
	}
}

func TestLogin_Success(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()
	s := NewServer(clientRegistry, gameRegistry)

	loginReq := &proto.LoginRequest{PlayerName: TEST_PLAYER_NAME}
	loginResp, err := s.Login(context.Background(), loginReq)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if loginResp.ErrorCode != proto.LoginResponse_UNSPECIFIED_ERR {
		t.Fatalf("wrong error code")
	}

	if !clientRegistry.HasRecord(loginResp.Token) {
		t.Fatalf("client not in registry!")
	}
}

func TestLogin_NameTaken(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()
	s := NewServer(clientRegistry, gameRegistry)

	clientRegistry.RegisterClient(TEST_TOKEN, TEST_PLAYER_NAME)

	request := &proto.LoginRequest{PlayerName: TEST_PLAYER_NAME}
	response, err := s.Login(context.Background(), request)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if response.ErrorCode != proto.LoginResponse_NAME_TAKEN_ERR {
		t.Fatalf("wrong error code")
	}

	if response.Token != "" {
		t.Fatalf("expected no token")
	}
}

func TestCreateGame(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()
	server := NewServer(clientRegistry, gameRegistry)

	clientRegistry.RegisterClient(TEST_TOKEN, TEST_PLAYER_NAME)

	ctx := getAuthContext(TEST_TOKEN)

	gameReq := &proto.CreateGameRequest{
		GameConfig: &proto.GameConfig{
			MaxPlayers:   5,
			Rounds:       10,
			RoundSeconds: 300,
		},
	}

	gameResp, _ := server.CreateGame(ctx, gameReq)

	if gameResp.GameCode == "" {
		t.Errorf("response does not have game code")
	}

	ok := gameRegistry.HasRecord(gameResp.GameID)

	if !ok {
		t.Errorf("game not in registry!")
	}
}
