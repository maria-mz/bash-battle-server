package game

import (
	"errors"
)

type GameStatus int
type GameScore int

type RoundNumber int

const (
	GAME_NOT_STARTED GameStatus = 0
	GAME_IN_PROGRESS GameStatus = 1
	GAME_OVER        GameStatus = 2
)

type GameConfig struct {
	Plan         GamePlan
	RoundSeconds int
}

type GameMeta struct {
	Status       GameStatus
	CurrentRound RoundNumber
}

// GameStore contains information about a Bash Battle game.
// Does little logic, up to the caller to use it well.
type GameStore struct {
	config  GameConfig
	meta    GameMeta
	players map[PlayerName]*Player
}

func NewGameStore(config GameConfig) *GameStore {
	return &GameStore{
		config:  config,
		players: make(map[PlayerName]*Player),
	}
}

// GetGameStatus gets the game's current status.
func (store *GameStore) GetGameStatus() GameStatus {
	return store.meta.Status
}

// SetGameStatus sets the game's status.
func (store *GameStore) SetGameStatus(status GameStatus) {
	store.meta.Status = status
}

// GetNumberOfRounds gets the configured number of rounds.
func (store *GameStore) GetNumRounds() int {
	return store.config.Plan.GetNumRounds()
}

// GetRoundNumber gets the current round number.
func (store *GameStore) GetRoundNumber() RoundNumber {
	return store.meta.CurrentRound
}

// SetRoundNumber sets the current round number.
func (store *GameStore) SetRoundNumber(num RoundNumber) {
	store.meta.CurrentRound = num
}

// AddPlayer adds a new player to the store with initial stats.
// Returns an error if there is already a player with the same name.
func (store *GameStore) AddPlayer(name PlayerName) error {
	_, ok := store.players[name]

	if ok {
		return errors.New("already a player with this name")
	}

	player := NewPlayer(name)
	store.players[name] = player

	return nil
}

// DeletePlayer deletes a player and all their info from the store.
func (store *GameStore) DeletePlayer(name PlayerName) {
	delete(store.players, name)
}

// GetPlayerScore gets a player's game score, and a success flag.
func (store *GameStore) GetPlayerScore(name PlayerName) (GameScore, bool) {
	player, ok := store.players[name]

	if !ok {
		return 0, false
	}

	return player.Stats.Score, true
}

// GetPlayerRoundStat gets the player's RoundStat for a particular round, and
// a success flag.
func (store *GameStore) GetPlayerRoundStat(
	name PlayerName,
	num RoundNumber,
) (RoundStat, bool) {
	player, ok := store.players[name]

	if !ok {
		return RoundStat{}, false
	}

	info, ok := player.Stats.RoundStats[num]
	return *info, ok
}

// SetPlayerRoundStat sets the player's RoundStat for a particular round.
// If the player isn't found, set is a no op and an error is returned.
func (store *GameStore) SetPlayerRoundStat(
	name PlayerName,
	num RoundNumber,
	stat RoundStat,
) error {
	player, ok := store.players[name]

	if !ok {
		return errors.New("player not found")
	}

	player.Stats.RoundStats[num] = &stat

	return nil
}

// GetRoundInfo gets the RoundInfo for a particular round, and a success flag.
func (store *GameStore) GetRoundInfo(num RoundNumber) (RoundInfo, bool) {
	info, ok := store.config.Plan.GetRoundInfo(num)
	return info, ok
}
