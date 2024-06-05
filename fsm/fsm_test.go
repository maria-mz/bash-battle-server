// Tests for Game Finite State Machine.
// If ANY test times out, something went wrong.
package fsm

import (
	"testing"
	"time"

	"github.com/maria-mz/bash-battle-server/log"
	"github.com/stretchr/testify/assert"
)

var testConfigMultiPlayer = FSMConfig{
	MaxPlayers:    3,
	Rounds:        2,
	RoundDuration: 3 * time.Second,
}

var testConfigSinglePlayer = FSMConfig{
	MaxPlayers:    1,
	Rounds:        2,
	RoundDuration: 3 * time.Second,
}

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

func TestNewGameFSM(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigMultiPlayer, updates)

	assert.NotNil(t, fsm)
	assert.Equal(t, WaitingForJoins, fsm.state)
	assert.Equal(t, 0, fsm.players.Size())
	assert.Equal(t, 0, fsm.confirms.Size())
	assert.Equal(t, 0, fsm.round)
}

func TestPlayerJoined(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigMultiPlayer, updates)

	fsm.PlayerJoined("player-1")

	assert.Equal(t, 1, fsm.players.Size())
	assert.Equal(t, WaitingForJoins, fsm.state)

	fsm.PlayerJoined("player-2")
	fsm.PlayerJoined("player-3")

	assert.Equal(t, 3, fsm.players.Size())
	assert.Equal(t, CountingDown, <-updates)
}

func TestPlayerLeft(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigMultiPlayer, updates)

	fsm.PlayerJoined("player-1")
	fsm.PlayerJoined("player-2")

	fsm.PlayerLeft("player-2")

	assert.Equal(t, 1, fsm.players.Size())
	assert.Equal(t, WaitingForJoins, fsm.state)

	fsm.PlayerLeft("player-1")

	assert.Equal(t, 0, fsm.players.Size())
	assert.Equal(t, WaitingForJoins, fsm.state)
}

func TestTypicalGameFlow(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigMultiPlayer, updates)

	fsm.PlayerJoined("player-1")
	fsm.PlayerJoined("player-2")
	fsm.PlayerJoined("player-3")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.PlayerConfirmed("player-1")
	fsm.PlayerConfirmed("player-2")
	fsm.PlayerConfirmed("player-3")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, Done, <-updates)
}

func TestCountingDownToWaitingForJoins(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigSinglePlayer, updates)

	fsm.PlayerJoined("player-1")

	assert.Equal(t, CountingDown, <-updates)

	time.Sleep(CountdownDuration / 2)

	fsm.PlayerLeft("player-1")

	assert.Equal(t, WaitingForJoins, <-updates)
}

func TestGameAbandonedDuringPlayingRound(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigSinglePlayer, updates)

	fsm.PlayerJoined("player-1")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)

	fsm.PlayerLeft("player-1")

	assert.Equal(t, Terminated, <-updates)
}

func TestGameAbandonedWhileWaitingForConfirms(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigSinglePlayer, updates)

	fsm.PlayerJoined("player-1")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.PlayerLeft("player-1")

	assert.Equal(t, Terminated, <-updates)
}

func TestGameAbandonedDuringCountdown(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigSinglePlayer, updates)

	fsm.PlayerJoined("player-1")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.PlayerConfirmed("player-1")

	assert.Equal(t, CountingDown, <-updates)

	fsm.PlayerLeft("player-1")

	assert.Equal(t, Terminated, <-updates)
}

func TestConfirmsWithLeave(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewGameFSM(testConfigMultiPlayer, updates)

	fsm.PlayerJoined("player-1")
	fsm.PlayerJoined("player-2")
	fsm.PlayerJoined("player-3")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.PlayerConfirmed("player-1")
	fsm.PlayerConfirmed("player-2")

	assert.Equal(t, 2, fsm.confirms.Size())

	// This player contributed a confirm
	fsm.PlayerLeft("player-1")

	// So, confirms should go down by 1
	assert.Equal(t, 1, fsm.confirms.Size())

	// This player has not confirmed
	// Them leaving should allow next round to start
	fsm.PlayerLeft("player-3")

	assert.Equal(t, 1, fsm.players.Size())

	// Game should continue to next round
	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, Done, <-updates)
}
