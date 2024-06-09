package data

import (
	"errors"
	"time"

	"github.com/maria-mz/bash-battle-server/config"
)

var ErrTooManyPlayers error = errors.New("max players already reached")
var ErrDuplicatePlayer error = errors.New("a player with this ID already exists")
var ErrPlayerNotFound error = errors.New("a player with this ID does not exist")
var ErrInvalidRound error = errors.New("round number is invalid")

type GameData struct {
	Config     config.GameConfig
	Challenges map[int]Challenge

	players map[string]*Player
}

func NewGameData(config config.GameConfig) *GameData {
	return &GameData{
		Config:     config,
		Challenges: GenerateChallenges(config),
		players:    make(map[string]*Player),
	}
}

func (data *GameData) HasPlayer(playerID string) bool {
	_, ok := data.players[playerID]
	return ok
}

func (data *GameData) GetPlayer(playerID string) (Player, bool) {
	player, ok := data.players[playerID]
	return *player, ok
}

func (data *GameData) NumPlayers() int {
	return len(data.players)
}

func (data *GameData) AddPlayer(playerID string, name string) error {
	if data.IsGameFull() {
		return ErrTooManyPlayers
	}

	_, ok := data.players[playerID]

	if ok {
		return ErrDuplicatePlayer
	}

	data.players[playerID] = NewPlayer(playerID, name)
	return nil
}

func (data *GameData) RemovePlayer(playerID string) {
	delete(data.players, playerID)
}

func (data *GameData) SetPlayerScore(playerID string, score Score) error {
	player, ok := data.players[playerID]

	if !ok {
		return ErrPlayerNotFound
	}

	if !data.IsRoundValid(score.Round) {
		return ErrInvalidRound
	}

	player.Scores[score.Round] = score
	return nil
}

func (data *GameData) GetChallenge(round int) (Challenge, error) {
	challenge, ok := data.Challenges[round]

	if !ok {
		return Challenge{}, ErrInvalidRound
	}

	return challenge, nil
}

func (data *GameData) GetRoundDuration() time.Duration {
	return time.Duration(data.Config.RoundDuration) * time.Second
}

func (data *GameData) GetCountdownDuration() time.Duration {
	return time.Duration(data.Config.CountdownDuration) * time.Second
}

func (data *GameData) IsGameFull() bool {
	return len(data.players) == data.Config.MaxPlayers
}

func (data *GameData) IsGameEmpty() bool {
	return len(data.players) == 0
}

func (data *GameData) IsRoundValid(round int) bool {
	return round < 1 || round > data.Config.Rounds
}
