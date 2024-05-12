package registry

import (
	"fmt"

	"github.com/maria-mz/bash-battle-server/game"
	id "github.com/maria-mz/bash-battle-server/idgen"
)

type ErrGameNotFound struct {
	GameID string
}

func (e ErrGameNotFound) Error() string {
	return fmt.Sprintf("no game record found with game ID %s", e.GameID)
}

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
	gameID := id.GenerateGameID()
	gameCode := id.GenerateGameCode()
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
func (registry *GameRegistry) GetGameCode(gameID string) (string, bool) {
	record, ok := registry.records[gameID]
	return record.GameCode, ok
}

// AddPlayer adds a new player to the game identified by the given game ID.
// It returns an error if the game does not exist.
func (registry *GameRegistry) AddPlayer(gameID string, name string) error {
	record, ok := registry.getRecord(gameID)

	if !ok {
		return ErrGameNotFound{gameID}
	}

	record.GameStore.AddPlayer(name)
	return nil
}

// getRecord retrieves the game record for the given game ID, if it exists.
func (registry *GameRegistry) getRecord(gameID string) (*GameRecord, bool) {
	record, ok := registry.records[gameID]
	return record, ok
}
