package game

type PlayerName string

type Player struct {
	Name  PlayerName
	Stats *PlayerStats
}

type PlayerStats struct {
	Score      GameScore
	RoundStats map[RoundNumber]*RoundStat
}

type RoundStat struct {
	WasBeat bool
	Command string
}

func NewPlayer(name PlayerName) *Player {
	stats := &PlayerStats{
		RoundStats: make(map[RoundNumber]*RoundStat),
	}
	return &Player{
		Name:  name,
		Stats: stats,
	}
}
