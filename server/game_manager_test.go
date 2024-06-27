package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGameManager(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.network)
	assert.NotNil(t, manager.gameData)
	assert.NotNil(t, manager.gameRunner)
	assert.Equal(t, manager.state, Lobby)
}

func TestAddClient_Normal(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	c1 := &client{username: "player-1", active: true}

	err := manager.AddClient(c1)

	assert.Nil(t, err)
	assert.Equal(t, manager.state, Lobby)
	assert.True(t, manager.gameData.HasPlayer("player-1"))
}

func TestAddClient_GameBecomesFull(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	c1 := &client{username: "player-1", active: true}
	c2 := &client{username: "player-2", active: true}
	c3 := &client{username: "player-3", active: true}

	manager.AddClient(c1)
	manager.AddClient(c2)
	manager.AddClient(c3)

	assert.Equal(t, manager.state, Load)
}

func TestAddClient_ErrJoinOnGameStarted(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	c1 := &client{username: "player-1", active: true}
	c2 := &client{username: "player-2", active: true}
	c3 := &client{username: "player-3", active: true}
	c4 := &client{username: "player-4", active: true}

	manager.AddClient(c1)
	manager.AddClient(c2)
	manager.AddClient(c3)
	err := manager.AddClient(c4)

	assert.Equal(t, err, ErrJoinOnGameStarted)
}

func TestAddClient_ErrUsernameTaken(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	c1 := &client{username: "player-1", active: true}
	c2 := &client{username: "player-1", active: true}

	manager.AddClient(c1)
	err := manager.AddClient(c2)

	assert.Equal(t, ErrUsernameTaken, err)
}
