package game

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maria-mz/bash-battle-server/log"
)

type RunnerEvent int

const (
	// CountingDown - Runner started counting down to the next round.
	CountingDown RunnerEvent = iota

	// RoundStarted - Timer started for the current round.
	RoundStarted

	// RoundEnded - Timer expired for the current round.
	RoundEnded
)

var ErrNoRoundsLeft error = errors.New("there are no rounds left to play")

// GameRunner runs the rounds of a Bash Battle game.
type GameRunner struct {
	GameData *GameData
	round    int
	ch       chan<- RunnerEvent
	mu       sync.Mutex
}

func NewGameRunner(game *GameData) (*GameRunner, <-chan RunnerEvent) {
	ch := make(chan RunnerEvent)

	runner := &GameRunner{
		GameData: game,
		ch:       ch,
	}

	return runner, ch
}

func (runner *GameRunner) GetCurrentRound() int {
	return runner.round
}

func (runner *GameRunner) RunRound() error {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	if runner.IsFinalRound() {
		return ErrNoRoundsLeft
	}

	runner.round++

	go runner.run()

	return nil
}

func (runner *GameRunner) run() {
	runner.ch <- CountingDown

	log.Logger.Info(fmt.Sprintf("Counting down to round %d", runner.round))
	runner.wait(runner.GameData.GetCountdownDuration())

	runner.ch <- RoundStarted

	log.Logger.Info(fmt.Sprintf("Started round %d", runner.round))
	runner.wait(runner.GameData.GetRoundDuration())

	runner.ch <- RoundEnded
	log.Logger.Info(fmt.Sprintf("Ended round %d", runner.round))

	if runner.IsFinalRound() {
		close(runner.ch)
		log.Logger.Info("Final round done. Closed RunnerEvent channel")
	}
}

func (runner *GameRunner) IsFinalRound() bool {
	return runner.round == runner.GameData.Config.Rounds
}

func (runner *GameRunner) wait(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
}
