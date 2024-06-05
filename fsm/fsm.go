package fsm

import (
	"fmt"
	"sync"
	"time"

	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
)

const CountdownDuration = 2 * time.Second

// GameFSM manages the state and synchronization logic for a Bash Battle game.
//
// The FSM responds to player actions and time-based events to transition
// between states. See docs/fsm-diagram.png for a diagram showing all the
// possible states and transitions.
//
// Player actions; the following methods (may) trigger state changes:
//   - PlayerJoined: A player has joined the game.
//   - PlayerLeft: A player has left the game.
//   - PlayerConfirmed: A player has confirmed they are ready for the next round.
//
// State changes are communicated through the provided FSMState channel.
type GameFSM struct {
	config FSMConfig

	state FSMState

	players  utils.Set[string]
	confirms utils.Set[string]
	round    int

	timer      *time.Timer
	cancelTime chan bool

	updates chan<- FSMState

	mutex sync.Mutex
}

func NewGameFSM(config FSMConfig, updates chan<- FSMState) *GameFSM {
	return &GameFSM{
		config:     config,
		cancelTime: make(chan bool),
		players:    utils.NewSet[string](),
		confirms:   utils.NewSet[string](),
		updates:    updates,
	}
}

func (fsm *GameFSM) PlayerJoined(name string) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.state != WaitingForJoins {
		return
	}

	if fsm.gameFull() {
		return
	}

	if fsm.players.Contains(name) {
		return
	}

	fsm.players.Add(name)

	log.Logger.Debug(
		"A new player joined the game",
		"name", name,
		"players", fsm.players.Size(),
	)

	if fsm.gameFull() {
		log.Logger.Info("Game is full, starting game!")
		go fsm.runNextRound()
	}
}

func (fsm *GameFSM) PlayerLeft(name string) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.gameEnded() {
		return
	}

	if fsm.gameEmpty() {
		return
	}

	if !fsm.players.Contains(name) {
		return
	}

	fsm.players.Delete(name)
	fsm.confirms.Delete(name)

	log.Logger.Debug(
		"An existing player left the game",
		"name", name,
		"players", fsm.players.Size(),
	)

	if fsm.isCountingDownToFirstRound() {
		log.Logger.Info("Game is no longer full, waiting for joins...")
		fsm.cancelTime <- true
		fsm.updateState(WaitingForJoins)
		return
	}

	if fsm.gameAbandoned() {
		log.Logger.Info("Game has been abandoned, terminating")
		fsm.endGame(Terminated)
		return
	}

	if fsm.state == WaitingForConfirms {
		fsm.checkConfirms()
		return
	}
}

func (fsm *GameFSM) PlayerConfirmed(name string) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.state != WaitingForConfirms {
		return
	}

	if !fsm.players.Contains(name) {
		return
	}

	fsm.confirms.Add(name)

	fsm.checkConfirms()
}

func (fsm *GameFSM) checkConfirms() {
	if fsm.confirms.Size() == fsm.players.Size() {
		log.Logger.Info("Every player has confirmed, starting next round")
		fsm.resetConfirms()
		go fsm.runNextRound()
	}
}

func (fsm *GameFSM) runNextRound() {
	fsm.round++

	fsm.updateState(CountingDown)

	if fsm.countdown() {
		fsm.updateState(PlayingRound)

		if fsm.playRound() {
			if fsm.isLastRound() {
				fsm.endGame(Done)
			} else {
				fsm.updateState(WaitingForConfirms)
			}
		}
	}
}

func (fsm *GameFSM) countdown() bool {
	fsm.timer = time.NewTimer(CountdownDuration)

	log.Logger.Info(fmt.Sprintf("Countdown to Round %d started", fsm.round))

	for {
		select {
		case <-fsm.cancelTime:
			log.Logger.Info(fmt.Sprintf("Countdown to Round %d cancelled", fsm.round))
			return false
		case <-fsm.timer.C:
			log.Logger.Info(fmt.Sprintf("Countdown to Round %d completed", fsm.round))
			return true
		}
	}
}

func (fsm *GameFSM) playRound() bool {
	fsm.timer = time.NewTimer(fsm.config.RoundDuration)

	log.Logger.Info(fmt.Sprintf("Round %d started", fsm.round))

	for {
		select {
		case <-fsm.cancelTime:
			log.Logger.Info(fmt.Sprintf("Round %d cancelled", fsm.round))
			return false
		case <-fsm.timer.C:
			log.Logger.Info(fmt.Sprintf("Round %d completed", fsm.round))
			return true
		}
	}
}

func (fsm *GameFSM) updateState(state FSMState) {
	log.Logger.Info(
		fmt.Sprintf(
			"FSM changes state: %s -> %s",
			stateValueMap[fsm.state],
			stateValueMap[state],
		),
	)

	fsm.state = state
	fsm.updates <- state
}

func (fsm *GameFSM) endGame(withState FSMState) {
	fsm.updateState(withState)
	close(fsm.cancelTime)
	close(fsm.updates)
}

func (fsm *GameFSM) isCountingDownToFirstRound() bool {
	return fsm.state == CountingDown && fsm.round == 1
}

func (fsm *GameFSM) gameFull() bool {
	return fsm.players.Size() == fsm.config.MaxPlayers
}

func (fsm *GameFSM) gameEmpty() bool {
	return fsm.players.Size() == 0
}

func (fsm *GameFSM) gameAbandoned() bool {
	return fsm.players.Size() == 0 && fsm.state != WaitingForJoins
}

func (fsm *GameFSM) gameEnded() bool {
	return fsm.state == Done || fsm.state == Terminated
}

func (fsm *GameFSM) isLastRound() bool {
	return fsm.round == fsm.config.Rounds
}

func (fsm *GameFSM) resetConfirms() {
	fsm.confirms = utils.NewSet[string]()
}
