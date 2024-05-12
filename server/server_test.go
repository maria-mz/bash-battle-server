package server

import (
	"context"
	"testing"

	"github.com/maria-mz/bash-battle-proto/proto"
	reg "github.com/maria-mz/bash-battle-server/registry"
)

func TestLogin_Success(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()

	server := NewServer(clientRegistry, gameRegistry)

	request := &proto.LoginRequest{PlayerName: "test-player-name"}

	response, err := server.Login(context.Background(), request)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if response.ErrorCode != proto.LoginResponse_UNSPECIFIED_ERR {
		t.Fatalf("expected no error code")
	}

	// Check that a record was actually added
	if !clientRegistry.HasRecord(response.Token) {
		t.Fatalf("expected client record to be in the registry")
	}
}

func TestLogin_NameTaken(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()
	gameRegistry := reg.NewGameRegistry()

	server := NewServer(clientRegistry, gameRegistry)

	request := &proto.LoginRequest{PlayerName: "test-player-name"}

	server.Login(context.Background(), request)
	response, err := server.Login(context.Background(), request)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if response.ErrorCode != proto.LoginResponse_NAME_TAKEN_ERR {
		t.Fatalf("expected name taken error code")
	}

	if response.Token != "" {
		t.Fatalf("expected no token")
	}
}
