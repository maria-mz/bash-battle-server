package main

import (
	"context"
	"testing"

	"github.com/maria-mz/bash-battle-proto/proto"
	reg "github.com/maria-mz/bash-battle-server/registry"
)

func TestLogin_Success(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()

	server := NewServer(clientRegistry)

	request := &proto.LoginRequest{Name: "test-player-name"}

	response, err := server.Login(context.Background(), request)
	t.Logf("response = %+v", response)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if response.Status != proto.LoginStatus_LoginSuccess {
		t.Fatalf("expected success login status")
	}

	// Check that a record was actually added
	if !clientRegistry.HasRecord(reg.ClientID(response.Token)) {
		t.Fatalf("expected client record to be in the registry")
	}
}

func TestLogin_NameTaken(t *testing.T) {
	clientRegistry := reg.NewClientRegistry()

	server := NewServer(clientRegistry)

	request := &proto.LoginRequest{Name: "test-player-name"}

	server.Login(context.Background(), request)
	response, err := server.Login(context.Background(), request)

	t.Logf("response = %+v", response)

	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if response.Status != proto.LoginStatus_NameTaken {
		t.Fatalf("expected name taken login status")
	}

	if response.Token != "" {
		t.Fatalf("expected no token")
	}
}
