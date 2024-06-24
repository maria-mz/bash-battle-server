package game

import (
	"testing"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/stretchr/testify/assert"
)

var testConfig = config.GameConfig{
	Rounds:            2,
	RoundDuration:     1, // seconds
	CountdownDuration: 1, // seconds
}

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

func TestNewGameRunner(t *testing.T) {
	data := NewGameData(testConfig)
	runner, _ := NewGameRunner(data)

	assert.NotNil(t, runner)
	assert.NotNil(t, runner.GameData)
	assert.NotNil(t, runner.ch)
	assert.Equal(t, 0, runner.GetCurrentRound())
}

func TestRunRound(t *testing.T) {
	data := NewGameData(testConfig)
	runner, ch := NewGameRunner(data)

	// first round - ok
	err := runner.RunRound()
	assert.Nil(t, err)

	assert.Equal(t, CountingDown, <-ch)
	assert.Equal(t, 1, runner.GetCurrentRound())
	assert.Equal(t, RoundStarted, <-ch)
	assert.Equal(t, 1, runner.GetCurrentRound())
	assert.Equal(t, RoundEnded, <-ch)
	assert.Equal(t, 1, runner.GetCurrentRound())

	// second (and last) round - ok
	err = runner.RunRound()
	assert.Nil(t, err)

	assert.Equal(t, CountingDown, <-ch)
	assert.Equal(t, 2, runner.GetCurrentRound())
	assert.Equal(t, RoundStarted, <-ch)
	assert.Equal(t, 2, runner.GetCurrentRound())
	assert.Equal(t, GameDone, <-ch)
	assert.Equal(t, 2, runner.GetCurrentRound())

	// error
	err = runner.RunRound()
	assert.NotNil(t, err)
	assert.Equal(t, err, ErrNoRoundsLeft)
}
