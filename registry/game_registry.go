package registry

import "github.com/maria-mz/bash-battle-server/game"

// -- Game record

type GameID string

type GameRecord struct {
	GameID    GameID
	GameStore *game.GameStore
	GameCode  string
}
