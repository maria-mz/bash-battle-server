package game

type FilePath string

type Round struct {
	Question   string
	InputFile  FilePath
	OutputFile FilePath
}

type GamePlan struct {
	rounds    map[int]Round
	numRounds int
}

func NewGamePlan() GamePlan {
	return GamePlan{
		rounds: make(map[int]Round),
	}
}

func (plan *GamePlan) AddRound(round Round) {
	plan.numRounds++
	plan.rounds[int(plan.numRounds)] = round
}

func (plan *GamePlan) GetNumRounds() int {
	return plan.numRounds
}

func (plan *GamePlan) GetRoundInfo(roundNumber int) (Round, bool) {
	info, ok := plan.rounds[roundNumber]
	return info, ok
}

// TODO: temporary, implement real functionality, randomly assign rounds
func BuildTempGamePlan(numRounds int) GamePlan {
	plan := NewGamePlan()

	info := Round{
		Question:   "???",
		InputFile:  "input.txt",
		OutputFile: "output.txt",
	}

	for i := 0; i < numRounds; i++ {
		plan.AddRound(info)
	}

	return plan
}
