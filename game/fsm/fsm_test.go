// Tests for Game Finite State Machine.
// If ANY test times out, something went wrong.
package fsm

import (
	"testing"
	"time"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/stretchr/testify/assert"
)

var testConfigMultiPlayer = config.GameConfig{
	MaxPlayers:        3,
	Rounds:            2,
	RoundDuration:     3, // seconds
	CountdownDuration: 2, // seconds
}

var testConfigSinglePlayer = config.GameConfig{
	MaxPlayers:        1,
	Rounds:            2,
	RoundDuration:     3, // seconds
	CountdownDuration: 2, // seconds
}

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

func TestNewGameFSM(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigMultiPlayer, updates)

	assert.NotNil(t, fsm)
	assert.Equal(t, WaitingForJoins, fsm.state)
	assert.Equal(t, 0, fsm.players.Size())
	assert.Equal(t, 0, fsm.confirms.Size())
	assert.Equal(t, 0, fsm.round)
}

func TestPlayerJoined(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigMultiPlayer, updates)

	fsm.AddPlayer("player-1")

	assert.Equal(t, 1, fsm.players.Size())
	assert.Equal(t, WaitingForJoins, fsm.state)

	fsm.AddPlayer("player-2")
	fsm.AddPlayer("player-3")

	assert.Equal(t, 3, fsm.players.Size())
	assert.Equal(t, CountingDown, <-updates)
}

func TestPlayerLeft(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigMultiPlayer, updates)

	fsm.AddPlayer("player-1")
	fsm.AddPlayer("player-2")

	fsm.RemovePlayer("player-2")

	assert.Equal(t, 1, fsm.players.Size())
	assert.Equal(t, WaitingForJoins, fsm.state)

	fsm.RemovePlayer("player-1")

	assert.Equal(t, 0, fsm.players.Size())
	assert.Equal(t, WaitingForJoins, fsm.state)
}

func TestTypicalGameFlow(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigMultiPlayer, updates)

	fsm.AddPlayer("player-1")
	fsm.AddPlayer("player-2")
	fsm.AddPlayer("player-3")

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
	fsm := NewFSM(testConfigSinglePlayer, updates)

	fsm.AddPlayer("player-1")

	assert.Equal(t, CountingDown, <-updates)

	time.Sleep(time.Second)

	fsm.RemovePlayer("player-1")

	assert.Equal(t, WaitingForJoins, <-updates)
}

func TestGameAbandonedDuringPlayingRound(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigSinglePlayer, updates)

	fsm.AddPlayer("player-1")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)

	fsm.RemovePlayer("player-1")

	assert.Equal(t, Terminated, <-updates)
}

func TestGameAbandonedWhileWaitingForConfirms(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigSinglePlayer, updates)

	fsm.AddPlayer("player-1")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.RemovePlayer("player-1")

	assert.Equal(t, Terminated, <-updates)
}

func TestGameAbandonedDuringCountdown(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigSinglePlayer, updates)

	fsm.AddPlayer("player-1")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.PlayerConfirmed("player-1")

	assert.Equal(t, CountingDown, <-updates)

	fsm.RemovePlayer("player-1")

	assert.Equal(t, Terminated, <-updates)
}

func TestConfirmsWithLeave(t *testing.T) {
	updates := make(chan FSMState, 1)
	fsm := NewFSM(testConfigMultiPlayer, updates)

	fsm.AddPlayer("player-1")
	fsm.AddPlayer("player-2")
	fsm.AddPlayer("player-3")

	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, WaitingForConfirms, <-updates)

	fsm.PlayerConfirmed("player-1")
	fsm.PlayerConfirmed("player-2")

	assert.Equal(t, 2, fsm.confirms.Size())

	// This player contributed a confirm
	fsm.RemovePlayer("player-1")

	// So, confirms should go down by 1
	assert.Equal(t, 1, fsm.confirms.Size())

	// This player has not confirmed
	// Them leaving should allow next round to start
	fsm.RemovePlayer("player-3")

	assert.Equal(t, 1, fsm.players.Size())

	// Game should continue to next round
	assert.Equal(t, CountingDown, <-updates)
	assert.Equal(t, PlayingRound, <-updates)
	assert.Equal(t, Done, <-updates)
}
