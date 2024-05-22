package server

import "github.com/maria-mz/bash-battle-proto/proto"

type EndStreamMsg struct {
	info string
	err  error
}

type ClientRecord struct {
	Token     string
	Username  string
	GameStats *proto.GameStats
	Stream    proto.BashBattle_StreamServer
	EndStream chan EndStreamMsg
}

func (record ClientRecord) ID() string {
	return record.Token
}

func NewClientRecord(token string, username string) *ClientRecord {
	return &ClientRecord{
		Token:    token,
		Username: username,
		GameStats: &proto.GameStats{
			Score:      0,
			RoundStats: make(map[int32]*proto.RoundStats),
		},
		EndStream: make(chan EndStreamMsg),
	}
}
