package game

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maria-mz/bash-battle-server/log"
)

var ErrNoRoundsLeft error = errors.New("there are no rounds left to play")

type RunnerEvent int

const (
	// CountingDown - Runner started counting down to the next round.
	CountingDown RunnerEvent = iota

	// RoundStarted - Timer started for the current round.
	RoundStarted

	// RoundEnded - Timer expired for the current (but not last) round.
	RoundEnded

	// GameDone - Timer expired for the last round.
	GameDone
)

// GameRunner runs the rounds of a Bash Battle game.
type GameRunner struct {
	Game  *Game
	round int

	ch chan<- RunnerEvent

	mu sync.Mutex
}

func NewRunner(game *Game) (*GameRunner, <-chan RunnerEvent) {
	ch := make(chan RunnerEvent)

	runner := &GameRunner{
		Game: game,
		ch:   ch,
	}

	return runner, ch
}

func (runner *GameRunner) GetCurrentRound() int {
	return runner.round
}

func (runner *GameRunner) RunRound() error {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	if runner.isFinalRound() {
		return ErrNoRoundsLeft
	}

	runner.round++

	go runner.run()

	return nil
}

func (runner *GameRunner) run() {
	runner.ch <- CountingDown

	log.Logger.Info(fmt.Sprintf("Counting down to round %d", runner.round))
	runner.wait(runner.Game.GetCountdownDuration())

	runner.ch <- RoundStarted

	log.Logger.Info(fmt.Sprintf("Started round %d", runner.round))
	runner.wait(runner.Game.GetRoundDuration())

	if runner.isFinalRound() {
		runner.ch <- GameDone

		log.Logger.Info(fmt.Sprintf("Ended round %d (final round)", runner.round))

		close(runner.ch)
		log.Logger.Info("Closed RunnerEvent channel")
	} else {
		runner.ch <- RoundEnded

		log.Logger.Info(fmt.Sprintf("Ended round %d", runner.round))
	}
}

func (runner *GameRunner) isFinalRound() bool {
	return runner.round == runner.Game.Config.Rounds
}

func (runner *GameRunner) wait(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
}
