package manager

import "github.com/maria-mz/bash-battle-server/utils"

type status int

const (
	// Joining - Waiting for players to join the game.
	Joining status = iota

	// Loading - Waiting for players to load the next round.
	Loading

	// Playing - Current round is playing.
	Playing

	// Submitting - Waiting for players to submit their score for the previous round.
	Submitting

	// Done - Game is complete!
	Done
)

type managerState struct {
	status           status
	playersWithLoad  utils.Set[string]
	playersWithScore utils.Set[string]
}

func initialManagerState() *managerState {
	return &managerState{
		status:           Joining,
		playersWithLoad:  utils.NewSet[string](),
		playersWithScore: utils.NewSet[string](),
	}
}

func (state *managerState) Status() status {
	return state.status
}

func (state *managerState) SetStatus(status status) {
	state.status = status
}

func (state *managerState) AddPlayerWithLoad(playerID string) {
	state.playersWithLoad.Add(playerID)
}

func (state *managerState) AddPlayerWithScore(playerID string) {
	state.playersWithScore.Add(playerID)
}

func (state *managerState) ResetPlayersWithLoad() {
	state.playersWithLoad = utils.NewSet[string]()
}

func (state *managerState) ResetPlayersWithScore() {
	state.playersWithScore = utils.NewSet[string]()
}

func (state *managerState) NumPlayersWithLoad() int {
	return state.playersWithLoad.Size()
}

func (state *managerState) NumPlayersWithScore() int {
	return state.playersWithScore.Size()
}
