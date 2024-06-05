package fsm

type FSMState int

const (
	WaitingForJoins FSMState = iota
	CountingDown
	PlayingRound
	WaitingForConfirms
	Terminated
	Done
)

// Maps State ints to string value. For logging purposes.
var stateValueMap = map[FSMState]string{
	WaitingForJoins:    "WAITING_FOR_JOINS",
	CountingDown:       "COUNTING_DOWN",
	PlayingRound:       "PLAYING_ROUND",
	WaitingForConfirms: "WAITING_FOR_CONFIRMS",
	Terminated:         "TERMINATED",
	Done:               "DONE",
}
