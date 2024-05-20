package game

import (
	"fmt"
	"time"

	reg "github.com/maria-mz/bash-battle-server/registry"
)

type GameState int

const (
	InLobby    GameState = 0
	InProgress GameState = 1
	Cancelled  GameState = 2
	Done       GameState = 3
)

type InvalidOp struct {
	Info string
}

func (err InvalidOp) Error() string {
	return err.Info
}

type GameConfig struct {
	RoundSeconds int
}

type Game struct {
	Config       GameConfig
	Plan         GamePlan
	State        GameState
	CurrentRound int
	Players      reg.Registry[string, Player]
	timer        time.Timer
	onRoundDone  func()
}

// NewGame creates an initial game with no players.
func NewGame(config GameConfig, plan GamePlan, onRoundDone func()) Game {
	return Game{
		Config:       config,
		Plan:         plan,
		State:        InLobby,
		CurrentRound: 0,
		Players:      *reg.NewRegistry[string, Player](),
		onRoundDone:  onRoundDone,
	}
}

func (game *Game) StartNextRound() error {
	if game.State == Cancelled || game.State == Done {
		return InvalidOp{"cannot start next round if game is over"}
	}

	go game.runRound()

	return nil
}

func (game *Game) runRound() {
	game.CurrentRound++

	game.timer = *time.NewTimer(game.getRoundDuration()) // starts timer!
	<-game.timer.C

	fmt.Printf("Round %d finished\n", game.CurrentRound)
	game.onRoundDone()

	if game.CurrentRound == game.Plan.GetNumRounds() {
		game.State = Done
	}
}

func (game *Game) getRoundDuration() time.Duration {
	return time.Duration(game.Config.RoundSeconds) * time.Second
}
