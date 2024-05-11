package main

import "github.com/maria-mz/bash-battle-server/game"

// -- Generic record interface

type ID string

type IdentifiableRecord interface {
	ID() ID
}

// -- Game record

type GameID string

type GameRecord struct {
	GameID    GameID
	GameStore *game.GameStore
	GameCode  string
}

func (record GameRecord) ID() ID {
	return ID(record.GameID)
}

// -- Client record

type ClientID string

type ClientRecord struct {
	ClientID     ClientID
	PlayerName   *string
	JoinedGameID *string
}

func (record ClientRecord) ID() ID {
	return ID(record.ClientID)
}
