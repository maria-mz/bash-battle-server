package game

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
