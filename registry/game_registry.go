package registry

import (
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/idgen"
)

// GameRecord represents a record of a game in the game registry.
type GameRecord struct {
	// GameID uniquely identifies the game.
	GameID string

	// GameStore holds the data and state of the game.
	GameStore *game.GameStore

	// GameCode is the code that players can use to join the game.
	GameCode string
}

// GameRegistry manages records of games on the server.
type GameRegistry struct {
	// records maps GameIDs to their corresponding GameRecord.
	records map[string]*GameRecord
}

// NewGameRegistry creates an empty GameRegistry
func NewGameRegistry() *GameRegistry {
	return &GameRegistry{
		records: make(map[string]*GameRecord),
	}
}

// RegisterGame creates a new game record with a unique ID and code.
// It returns the generated game ID and code.
func (registry *GameRegistry) RegisterGame(config game.GameConfig) (string, string) {
	gameID := idgen.GenerateGameID()
	gameCode := idgen.GenerateGameCode()
	store := game.NewGameStore(config)

	record := &GameRecord{
		GameID:    gameID,
		GameCode:  gameCode,
		GameStore: store,
	}

	registry.records[gameID] = record

	return gameID, gameCode
}

// GetGameCode returns the game code for a game, if the record exists.
func (registry *GameRegistry) GetGameCode(id string) (string, bool) {
	record, ok := registry.records[id]
	return record.GameCode, ok
}
