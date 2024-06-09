package runner

type RunnerEvent int

const (
	// CountingDown - Runner started counting down to the next round.
	CountingDown RunnerEvent = iota

	// RoundStarted - Timer started for the current round.
	RoundStarted

	// RoundEnded - Timer expired for the current (but not last) round.
	RoundEnded

	// GameDone - Timer expired for the last round.
	GameDone
)
