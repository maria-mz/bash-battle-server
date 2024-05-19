package server

import (
	"context"
	"testing"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	reg "github.com/maria-mz/bash-battle-server/registry"
)

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

	if err != st.expectedErr {
		t.Errorf(
			"err mismatch: actual != expected: %s != %s",
			err,
			st.expectedErr,
		)
	}

	if resp.ErrorCode != st.expectedResp.ErrorCode {
		t.Errorf(
			"err code mismatch: actual != expected: %s != %s",
			resp.ErrorCode,
			st.expectedResp.ErrorCode,
		)
	}

	if st.shouldFail {
		if resp.Token != "" {
			t.Errorf("token should be empty")
		}
		if clients.HasRecord(resp.Token) {
			t.Errorf("there should not be a new client in the registry")
		}
	} else {
		if resp.Token == "" {
			t.Errorf("token should not be empty")
		}
		if !clients.HasRecord(resp.Token) {
			t.Errorf("a new client should've been added to the registry")
		}
	}
}

var loginTests = []loginTest{
	{
		name:        "first client + ok",
		games:       []GameRecord{},
		clients:     []ClientRecord{},
		request:     &pb.LoginRequest{PlayerName: testPlayerName},
		expectedErr: nil,
		expectedResp: &pb.LoginResponse{
			ErrorCode: pb.LoginResponse_UNSPECIFIED_ERR,
		},
		shouldFail: false,
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
			ErrorCode: pb.LoginResponse_NAME_TAKEN_ERR,
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
