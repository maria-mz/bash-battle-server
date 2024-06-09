package data

type Score struct {
	Round   int
	Win     bool
	CmdUsed string
}

type Player struct {
	ID     string
	Name   string
	Scores map[int]Score
}

func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:     id,
		Name:   name,
		Scores: make(map[int]Score),
	}
}
