package fsm

import "time"

type FSMConfig struct {
	MaxPlayers        int
	Rounds            int
	RoundDuration     time.Duration
	CountdownDuration time.Duration
}
