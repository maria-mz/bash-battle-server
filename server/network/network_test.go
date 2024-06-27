package network

import (
	"testing"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/server/client"
	"github.com/maria-mz/bash-battle-server/server/stream"
	"github.com/maria-mz/bash-battle-server/utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	m.Run()
}

func TestNewNetwork(t *testing.T) {
	network, clientMsgs := NewNetwork()

	assert.NotNil(t, network)
	assert.NotNil(t, clientMsgs)
	assert.NotNil(t, network.clients)
	assert.NotNil(t, network.activityBitmap)
}

func TestAddClient_Ok(t *testing.T) {
	network, _ := NewNetwork()

	c1 := &client.Client{Username: "player-1"}

	err := network.AddClient(c1)

	assert.Nil(t, err)
	assert.Equal(t, network.clients[c1.Username], c1)
}

func TestAddClient_ErrNameTaken(t *testing.T) {
	network, _ := NewNetwork()

	c1 := &client.Client{Username: "player-1"}
	c2 := &client.Client{Username: "player-1"}

	network.AddClient(c1)
	err := network.AddClient(c2)

	assert.Equal(t, ErrUsernameTaken, err)
	assert.Equal(t, network.clients[c1.Username], c1) // Ensure ptr is perserved
}

func TestBroadcast(t *testing.T) {
	network, _ := NewNetwork()

	mss1 := utils.NewMockStreamServer()
	mss2 := utils.NewMockStreamServer()

	c1 := &client.Client{
		Username: "player-1",
		Stream:   stream.NewStream(mss1),
		Active:   true,
	}
	c2 := &client.Client{Username: "player-2"} // No stream
	c3 := &client.Client{
		Username: "player-3",
		Stream:   stream.NewStream(mss2),
		Active:   true,
	}

	network.AddClient(c1)
	network.AddClient(c2)
	network.AddClient(c3)

	somePlayer := game.NewPlayer("some-player")

	network.BroadcastPlayerJoin(somePlayer)

	// If test hangs here something went wrong
	assert.NotNil(t, <-mss1.RecievedEvents)
	assert.NotNil(t, <-mss2.RecievedEvents)
}

func TestClientAck(t *testing.T) {
	network, clientMsgs := NewNetwork()

	mss := utils.NewMockStreamServer()

	c := &client.Client{
		Username: "player-1",
		Stream:   stream.NewStream(mss),
		Active:   true,
	}

	go func() {
		ackMsg := &proto.AckMsg{
			Ack: &proto.AckMsg_RoundLoaded{
				RoundLoaded: &proto.RoundLoaded{},
			},
		}

		mss.AckMsgs <- ackMsg // Mock stream should return this value on Recv call

		clientMsg := <-clientMsgs

		assert.Equal(t, c.Username, clientMsg.Username)
		assert.Equal(t, ackMsg, clientMsg.Msg)
		assert.True(t, network.GetClientLoadStatus(c.Username))

		mss.Close() // Actually ends test goroutine
	}()

	network.AddClient(c)
	network.ListenForClientMsgs(c.Username) // Blocks until mss.Close() is called
}
