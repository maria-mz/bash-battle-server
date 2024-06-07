package game

import (
	"errors"
	"sync"
	"time"

	"github.com/maria-mz/bash-battle-server/config"
)

var ErrTooManyPlayers error = errors.New("max players already reached")
var ErrDuplicatePlayer error = errors.New("a player with this name already exists")
var ErrPlayerDoesNotExist error = errors.New("a player with this name does not exist")
var ErrInvalidRound error = errors.New("round number is invalid")

type Game struct {
	Config     config.GameConfig
	Challenges map[int]Challenge

	players map[string]*Player

	mu sync.Mutex
}

func NewGame(config config.GameConfig) *Game {
	return &Game{
		Config:     config,
		Challenges: GenerateChallenges(config),
		players:    make(map[string]*Player),
	}
}

func (game *Game) HasPlayer(name string) bool {
	_, ok := game.players[name]
	return ok
}

func (game *Game) GetPlayer(name string) (Player, bool) {
	player, ok := game.players[name]
	return *player, ok
}

func (game *Game) AddNewPlayer(name string) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.IsGameFull() {
		return ErrTooManyPlayers
	}

	_, ok := game.players[name]

	if ok {
		return ErrDuplicatePlayer
	}

	game.players[name] = NewPlayer(name)
	return nil
}

func (game *Game) RemovePlayer(name string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.players, name)
}

func (game *Game) SetPlayerScore(name string, score Score) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, ok := game.players[name]

	if !ok {
		return ErrPlayerDoesNotExist
	}

	if score.Round > game.Config.Rounds {
		return ErrInvalidRound
	}

	player.Scores[score.Round] = score
	return nil
}

func (game *Game) GetChallenge(round int) (Challenge, error) {
	challenge, ok := game.Challenges[round]

	if !ok {
		return Challenge{}, ErrInvalidRound
	}

	return challenge, nil
}

func (game *Game) GetRoundDuration() time.Duration {
	return time.Duration(game.Config.RoundDuration) * time.Second
}

func (game *Game) GetCountdownDuration() time.Duration {
	return time.Duration(game.Config.CountdownDuration) * time.Second
}

func (game *Game) IsGameFull() bool {
	return len(game.players) == game.Config.MaxPlayers
}

func (game *Game) IsGameEmpty() bool {
	return len(game.players) == 0
}
