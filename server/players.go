package server

import (
	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/registry"
)

type PlayerRegistry struct {
	players *registry.Registry[string, proto.Player]
}

func NewPlayerRegistry() *PlayerRegistry {
	return &PlayerRegistry{
		players: registry.NewRegistry[string, proto.Player](),
	}
}

func (r *PlayerRegistry) AddPlayer(name string) {
	emptyStats := &proto.GameStats{
		RoundStats: make(map[int32]*proto.RoundStats),
		Score:      0,
	}

	player := &proto.Player{
		Username: name,
		Stats:    emptyStats,
	}

	r.players.WriteRecord(name, player)
}

func (r *PlayerRegistry) DeletePlayer(name string) {
	r.players.DeleteRecord(name)
}

func (r *PlayerRegistry) GetPlayers() []*proto.Player {
	return r.players.Records()
}

func (r *PlayerRegistry) GetPlayer(name string) (*proto.Player, bool) {
	player, ok := r.players.GetRecord(name)
	return player, ok
}

func (r *PlayerRegistry) HasPlayer(name string) bool {
	ok := r.players.HasRecord(name)
	return ok
}
