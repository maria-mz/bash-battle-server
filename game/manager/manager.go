package manager

import (
	"errors"
	"sync"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game/data"
	"github.com/maria-mz/bash-battle-server/game/runner"
)

var ErrGameStarted error = errors.New("the game has already started")

type GameManager struct {
	GameData         *data.GameData
	gameRunner       *runner.GameRunner
	gameRunnerEvents <-chan runner.RunnerEvent
	cmds             chan<- GameManagerCmd
	state            *managerState
	mu               sync.Mutex
}

func NewGameManager(config config.GameConfig) (*GameManager, <-chan GameManagerCmd) {
	data := data.NewGameData(config)
	runner, runnerEvents := runner.NewGameRunner(data)

	mgrCmds := make(chan GameManagerCmd)

	gameMgr := &GameManager{
		GameData:         data,
		gameRunner:       runner,
		gameRunnerEvents: runnerEvents,
		cmds:             mgrCmds,
		state:            initialManagerState(),
	}

	go gameMgr.handleRunnerEvents()

	return gameMgr, mgrCmds
}

func (mgr *GameManager) handleRunnerEvents() {
EventLoop:
	for event := range mgr.gameRunnerEvents {
		round := mgr.gameRunner.GetCurrentRound()

		switch event {
		case runner.CountingDown:
			mgr.cmds <- SendRoundStartTime{round}

		case runner.RoundStarted:
			mgr.cmds <- SendRoundEndTime{round}

		case runner.RoundEnded:
			mgr.state.SetStatus(Submitting)
			mgr.cmds <- SubmitScore{round}

		case runner.GameDone:
			mgr.state.SetStatus(Done)
			close(mgr.cmds)

			break EventLoop
		}
	}
}

func (mgr *GameManager) AddPlayer(playerID string, name string) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if mgr.state.Status() != Joining {
		return ErrGameStarted
	}

	if err := mgr.GameData.AddPlayer(playerID, name); err != nil {
		return err
	}

	if mgr.GameData.IsGameFull() {
		mgr.state.SetStatus(Loading)
		mgr.cmds <- LoadRound{Round: mgr.gameRunner.GetCurrentRound() + 1}
	}

	return nil
}

// TODO: Implement RemovePlayer

func (mgr *GameManager) LoadedRoundFor(playerID string) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if mgr.state.Status() == Loading && mgr.GameData.HasPlayer(playerID) {
		mgr.state.AddPlayerWithLoad(playerID)
		mgr.checkLoads()
	}
}

func (mgr *GameManager) SubmitScoreFor(playerID string, score data.Score) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if mgr.state.Status() == Submitting && mgr.GameData.HasPlayer(playerID) {
		err := mgr.GameData.SetPlayerScore(playerID, score)

		if err != nil {
			return
		}

		mgr.state.AddPlayerWithScore(playerID)
		mgr.checkSubmissions()
	}
}

func (mgr *GameManager) checkLoads() {
	if mgr.state.NumPlayersWithLoad() != mgr.GameData.NumPlayers() {
		return
	}

	mgr.state.ResetPlayersWithLoad()
	mgr.state.SetStatus(Playing)

	mgr.gameRunner.RunRound()
}

func (mgr *GameManager) checkSubmissions() {
	if mgr.state.NumPlayersWithScore() != mgr.GameData.NumPlayers() {
		return
	}

	mgr.state.ResetPlayersWithScore()
	mgr.state.SetStatus(Loading)

	mgr.cmds <- LoadRound{Round: mgr.gameRunner.GetCurrentRound() + 1}
}
