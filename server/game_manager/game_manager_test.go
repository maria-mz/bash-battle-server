package game_manager

import (
	"testing"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server/network"
	"github.com/stretchr/testify/assert"
)

var testConfig = config.Config{
	GameConfig: config.GameConfig{
		MaxPlayers:        3,
		Rounds:            2,
		RoundDuration:     2,
		CountdownDuration: 1,
	},
}

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

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

	c1 := &network.Client{Username: "player-1"}

	err := manager.AddClient(c1)

	assert.Nil(t, err)
	assert.Equal(t, manager.state, Lobby)
	assert.True(t, manager.gameData.HasPlayer("player-1"))
}

func TestAddClient_GameBecomesFull(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	c1 := &network.Client{Username: "player-1"}
	c2 := &network.Client{Username: "player-2"}
	c3 := &network.Client{Username: "player-3"}

	manager.AddClient(c1)
	manager.AddClient(c2)
	manager.AddClient(c3)

	assert.Equal(t, Load, manager.state)
}

func TestAddClient_ErrJoinOnGameStarted(t *testing.T) {
	manager := NewGameManager(testConfig.GameConfig)

	c1 := &network.Client{Username: "player-1"}
	c2 := &network.Client{Username: "player-2"}
	c3 := &network.Client{Username: "player-3"}
	c4 := &network.Client{Username: "player-4"}

	manager.AddClient(c1)
	manager.AddClient(c2)
	manager.AddClient(c3)
	err := manager.AddClient(c4)

	assert.Equal(t, err, ErrJoinOnGameStarted)
}
