package game

type FilePath string

type RoundInfo struct {
	Question   string
	InputFile  FilePath
	OutputFile FilePath
}

type GamePlan struct {
	rounds    map[RoundNumber]RoundInfo
	numRounds int
}

func NewGamePlan() GamePlan {
	return GamePlan{
		rounds: make(map[RoundNumber]RoundInfo),
	}
}

func (plan *GamePlan) AddRound(round RoundInfo) {
	plan.numRounds++
	plan.rounds[RoundNumber(plan.numRounds)] = round
}

func (plan *GamePlan) GetNumRounds() int {
	return plan.numRounds
}

func (plan *GamePlan) GetRoundInfo(num RoundNumber) (RoundInfo, bool) {
	info, ok := plan.rounds[num]
	return info, ok
}

// TODO: temporary, implement real functionality, randomly assign rounds
func BuildTempGamePlan(rounds int) GamePlan {
	plan := NewGamePlan()

	info := RoundInfo{
		Question:   "???",
		InputFile:  "input.txt",
		OutputFile: "output.txt",
	}

	for i := 0; i < rounds; i++ {
		plan.AddRound(info)
	}

	return plan
}
