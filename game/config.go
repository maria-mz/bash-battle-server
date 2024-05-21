package game

type GameConfig struct {
	MaxPlayers   int
	Rounds       int
	RoundSeconds int
	Difficulty   Difficulty
	FileSize     FileSize
}
