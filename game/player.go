package game

type Player struct {
	Name  string
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

func NewPlayer(name string) *Player {
	stats := &PlayerStats{
		RoundStats: make(map[RoundNumber]*RoundStat),
	}
	return &Player{
		Name:  name,
		Stats: stats,
	}
}
