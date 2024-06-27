package game

import pb "github.com/maria-mz/bash-battle-proto/proto"

type Score struct {
	Round   int
	Win     bool
	CmdUsed string
}

type Player struct {
	Name   string
	Scores map[int]Score
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:   name,
		Scores: make(map[int]Score),
	}
}

func (player *Player) SetRoundScore(score Score) {
	player.Scores[score.Round] = score
}

// Don't know how much I like this but it is what it is
func (player *Player) ToProto() *pb.Player {
	gameStats := &pb.GameStats{
		RoundStats: make(map[int32]*pb.RoundStats),
	}

	for round, score := range player.Scores {
		roundStats := &pb.RoundStats{
			Won:     score.Win,
			Command: score.CmdUsed,
		}
		gameStats.RoundStats[int32(round)] = roundStats
	}

	return &pb.Player{
		Username: player.Name,
		Stats:    gameStats,
	}
}
