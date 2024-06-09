package runner

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maria-mz/bash-battle-server/game/data"
	"github.com/maria-mz/bash-battle-server/log"
)

var ErrNoRoundsLeft error = errors.New("there are no rounds left to play")

// GameRunner runs the rounds of a Bash Battle game.
type GameRunner struct {
	GameData *data.GameData
	round    int
	ch       chan<- RunnerEvent
	mu       sync.Mutex
}

func NewGameRunner(game *data.GameData) (*GameRunner, <-chan RunnerEvent) {
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
	runner.wait(runner.GameData.GetCountdownDuration())

	runner.ch <- RoundStarted

	log.Logger.Info(fmt.Sprintf("Started round %d", runner.round))
	runner.wait(runner.GameData.GetRoundDuration())

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
	return runner.round == runner.GameData.Config.Rounds
}

func (runner *GameRunner) wait(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
}
