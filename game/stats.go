package game

import "github.com/maria-mz/bash-battle-proto/proto"

func NewGameStats() *proto.GameStats {
	return &proto.GameStats{
		Score:      0,
		RoundStats: make(map[int32]*proto.RoundStats),
	}
}
