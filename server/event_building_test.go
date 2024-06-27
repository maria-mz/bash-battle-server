package server

import (
	"testing"
	"time"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/stretchr/testify/assert"
)

func TestBuildPlayerJoinedEvent(t *testing.T) {
	p := game.NewPlayer("player-1")
	s1 := game.Score{Round: 1, Win: true}

	p.SetRoundScore(s1)

	event := buildPlayerJoinedEvent(p)

	assert.NotNil(t, event)
	assert.NotNil(t, event.GetPlayerJoined())
	assert.Equal(t, event.GetPlayerJoined().GetPlayer(), p.ToProto())
}

func TestBuildPlayerLeftEvent(t *testing.T) {
	p := game.NewPlayer("player-1")
	s1 := game.Score{Round: 1, Win: true}

	p.SetRoundScore(s1)

	event := buildPlayerLeftEvent(p)

	assert.NotNil(t, event)
	assert.NotNil(t, event.GetPlayerLeft())
	assert.Equal(t, p.ToProto(), event.GetPlayerLeft().GetPlayer())
}

func TestBuildCountingDownEvent(t *testing.T) {
	round := 1
	startsAt := time.Now()

	event := buildCountingDownEvent(round, startsAt)

	assert.NotNil(t, event)
	assert.NotNil(t, event.GetCountingDown())
	assert.Equal(t, round, int(event.GetCountingDown().GetRoundNumber()))
	// NOTE: timestamppb.Timestamp.AsTime() gets time.Time in UTC time standard
	assert.Equal(t, startsAt.UTC(), event.GetCountingDown().GetStartsAt().AsTime())
}

func TestBuildRoundStartedEvent(t *testing.T) {
	round := 1
	endsAt := time.Now()

	event := buildRoundStartedEvent(round, endsAt)

	assert.NotNil(t, event)
	assert.NotNil(t, event.GetRoundStarted())
	assert.Equal(t, round, int(event.GetRoundStarted().GetRoundNumber()))
	// NOTE: timestamppb.Timestamp.AsTime() gets time.Time in UTC time standard
	assert.Equal(t, endsAt.UTC(), event.GetRoundStarted().GetEndsAt().AsTime())
}

func TestBuildLoadRoundEvent(t *testing.T) {
	round := 1
	challenge := game.Challenge{
		Question: "sample-question",
		// TODO: Add file paths and check bytes in pb message
	}

	event := buildLoadRoundEvent(round, challenge)

	assert.NotNil(t, event)
	assert.NotNil(t, event.GetLoadRound())
	assert.Equal(t, round, int(event.GetLoadRound().GetRoundNumber()))
	assert.Equal(t, challenge.Question, event.GetLoadRound().Question)
}

func TestBuildSubmitRoundScoreEvent(t *testing.T) {
	event := buildSubmitRoundScoreEvent()

	assert.NotNil(t, event)

	// Since message inside is empty assert.NotNil() will evaluate to false
	// Validate type by type assertion instead
	_, ok := event.GetEvent().(*proto.Event_SubmitRoundScore)
	assert.True(t, ok)
}

func TestBuildGameOverEvent(t *testing.T) {
	event := buildGameOverEvent()

	assert.NotNil(t, event)

	// Since message inside is empty assert.NotNil() will evaluate to false
	// Validate type by type assertion instead
	_, ok := event.GetEvent().(*proto.Event_GameOver)
	assert.True(t, ok)
}
