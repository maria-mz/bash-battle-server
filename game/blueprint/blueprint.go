package blueprint

import "github.com/maria-mz/bash-battle-server/config"

type FilePath string

type RoundInfo struct {
	Question   string
	InputFile  FilePath
	OutputFile FilePath
}

type Blueprint struct {
	rounds    map[int]RoundInfo
	numRounds int
}

func NewBlueprint() Blueprint {
	return Blueprint{
		rounds: make(map[int]RoundInfo),
	}
}

func (plan *Blueprint) AddRound(round RoundInfo) {
	plan.numRounds++
	plan.rounds[int(plan.numRounds)] = round
}

func (plan *Blueprint) GetNumRounds() int {
	return plan.numRounds
}

func (plan *Blueprint) GetRoundInfo(roundNumber int) (RoundInfo, bool) {
	info, ok := plan.rounds[roundNumber]
	return info, ok
}

// TODO: temporary, implement real functionality, randomly assign rounds
func BuildBlueprint(config config.GameConfig) Blueprint {
	plan := NewBlueprint()

	info := RoundInfo{
		Question:   "???",
		InputFile:  "input.txt",
		OutputFile: "output.txt",
	}

	for i := 0; i < config.Rounds; i++ {
		plan.AddRound(info)
	}

	return plan
}
