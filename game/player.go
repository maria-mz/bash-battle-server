package game

type Player struct {
	PlayerID string
	Name     string
	Score    int
	Stats    map[int]*RoundStat
}

func (p Player) ID() string {
	return p.PlayerID
}

type RoundStat struct {
	Won     bool
	Command string
}

func NewPlayer(id string, name string) Player {
	return Player{
		PlayerID: id,
		Name:     name,
		Stats:    make(map[int]*RoundStat),
	}
}

func (p *Player) UpdateStats(round int, stats *RoundStat) {
	p.Stats[round] = stats
}
