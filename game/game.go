package game

import (
	"fmt"
	"sync"
	"time"
)

type InvalidOp struct {
	Info string
}

func (err InvalidOp) Error() string {
	return err.Info
}

type Game struct {
	Config          GameConfig
	Plan            GamePlan
	State           GameState
	CurrentRound    int
	timer           time.Timer
	roundInProgress bool
	onRoundDone     func()
	mutex           sync.Mutex
}

// NewGame creates an initial game with no players.
func NewGame(config GameConfig, plan GamePlan, onRoundDone func()) *Game {
	return &Game{
		Config:       config,
		Plan:         plan,
		State:        InLobby,
		CurrentRound: 0,
		onRoundDone:  onRoundDone,
	}
}

func (game *Game) StartNextRound() error {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	if game.State == Cancelled || game.State == Done {
		return InvalidOp{"game is over; no rounds left to run"}
	}

	if game.roundInProgress {
		return InvalidOp{"round is currently in progress!"}
	}

	game.State = InProgress

	go game.runRound()

	return nil
}

func (game *Game) runRound() {
	game.roundInProgress = true

	game.CurrentRound++

	game.timer = *time.NewTimer(game.getRoundDuration()) // starts timer!
	<-game.timer.C

	fmt.Printf("Round %d finished\n", game.CurrentRound)
	game.onRoundDone()

	if game.CurrentRound == game.Plan.GetNumRounds() {
		game.State = Done
	}

	game.roundInProgress = false
}

func (game *Game) getRoundDuration() time.Duration {
	return time.Duration(game.Config.RoundSeconds) * time.Second
}
