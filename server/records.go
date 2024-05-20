package server

import (
	"github.com/maria-mz/bash-battle-server/game"
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
	GameID string
	Code   string
	Game   game.Game
}

func (record GameRecord) ID() string {
	return record.GameID
}
