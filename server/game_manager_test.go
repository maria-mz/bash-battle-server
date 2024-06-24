package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGameManager(t *testing.T) {
	gameManger := NewGameManager(testConfig.GameConfig)

	assert.NotNil(t, gameManger)
	assert.NotNil(t, gameManger.clients)
	assert.NotNil(t, gameManger.clientAckBitmap)
	assert.NotNil(t, gameManger.gameData)
	assert.NotNil(t, gameManger.gameRunner)
	assert.Equal(t, gameManger.state, Lobby)
	assert.Equal(t, len(gameManger.clients), 0)
}

func TestAddClient_Normal(t *testing.T) {
	gameManager := NewGameManager(testConfig.GameConfig)

	client1 := &client{username: "player-1", active: true}

	err := gameManager.AddClient(client1)

	assert.Nil(t, err)
	assert.Equal(t, gameManager.state, Lobby)
	assert.True(t, gameManager.gameData.HasPlayer("player-1"))
}

func TestAddClient_GameBecomesFull(t *testing.T) {
	gameManager := NewGameManager(testConfig.GameConfig)

	client1 := &client{username: "player-1", active: true}
	client2 := &client{username: "player-2", active: true}
	client3 := &client{username: "player-3", active: true}

	gameManager.AddClient(client1)
	gameManager.AddClient(client2)
	gameManager.AddClient(client3)

	assert.Equal(t, gameManager.state, Load)
}

func TestAddClient_ErrJoinOnGameStarted(t *testing.T) {
	gameManager := NewGameManager(testConfig.GameConfig)

	client1 := &client{username: "player-1", active: true}
	client2 := &client{username: "player-2", active: true}
	client3 := &client{username: "player-3", active: true}
	client4 := &client{username: "player-4", active: true}

	gameManager.AddClient(client1)
	gameManager.AddClient(client2)
	gameManager.AddClient(client3)
	err := gameManager.AddClient(client4)

	assert.Equal(t, err, ErrJoinOnGameStarted)
}

func TestAddClient_ErrUsernameTaken(t *testing.T) {
	gameManager := NewGameManager(testConfig.GameConfig)

	client1 := &client{username: "player-1", active: true}
	client2 := &client{username: "player-1", active: true}

	gameManager.AddClient(client1)
	err := gameManager.AddClient(client2)

	assert.Equal(t, err, ErrUsernameTaken)
}
