package game

type (
	GameState  int32
	Difficulty int32
	FileSize   int32
)

const (
	InLobby GameState = iota
	InProgress
	Cancelled
	Done
)

const (
	VariedDiff Difficulty = iota
	EasyDiff
	MediumDiff
	HardDiff
)

const (
	VariedSize FileSize = iota
	SmallSize
	MediumSize
	BigSize
)
