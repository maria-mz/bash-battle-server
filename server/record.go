package server

import "github.com/maria-mz/bash-battle-proto/proto"

type ClientRecord struct {
	Token     string
	Username  string
	GameStats *proto.GameStats
}

func (record ClientRecord) ID() string {
	return record.Token
}
