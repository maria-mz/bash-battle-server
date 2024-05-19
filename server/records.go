package server

import (
	"github.com/maria-mz/bash-battle-server/state"
	"github.com/maria-mz/bash-battle-server/utils"
)

type ClientRecord struct {
	ClientID   string
	PlayerName string
	GameID     *string
}

func (record ClientRecord) ID() string {
	return record.ClientID
}

type GameRecord struct {
	GameID    string
	GameStore *state.GameStore
	GameCode  string
	Members   utils.StrSet
}

func (record GameRecord) ID() string {
	return record.GameID
}
