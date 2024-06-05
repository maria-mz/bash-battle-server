package fsm

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/utils"
)

// FSM manages the state and synchronization logic for a Bash Battle game.
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
type FSM struct {
	config config.GameConfig

	state FSMState

	players  utils.Set[string]
	confirms utils.Set[string]
	round    int

	timer      *time.Timer
	cancelTime chan bool

	updates chan<- FSMState

	// This mutex ensures that only one player action is processed at time.
	mutex sync.Mutex
}

func NewFSM(config config.GameConfig, updates chan<- FSMState) *FSM {
	return &FSM{
		config:     config,
		cancelTime: make(chan bool),
		players:    utils.NewSet[string](),
		confirms:   utils.NewSet[string](),
		updates:    updates,
	}
}

func (fsm *FSM) AddPlayer(name string) error {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.state != WaitingForJoins {
		return errors.New("game is not accepting joins")
	}

	if fsm.isGameFull() {
		return errors.New("game is full")
	}

	if fsm.players.Contains(name) {
		return fmt.Errorf("already a player with name %s", name)
	}

	fsm.players.Add(name)

	log.Logger.Debug(
		"A new player joined the game",
		"name", name,
		"players", fsm.players.Size(),
	)

	if fsm.isGameFull() {
		log.Logger.Info("Game is full, starting game!")
		go fsm.runNextRound()
	}

	return nil
}

func (fsm *FSM) RemovePlayer(name string) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	if fsm.hasGameEnded() {
		return
	}

	if fsm.isGameEmpty() {
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

	if fsm.isGameAbandoned() {
		log.Logger.Info("Game has been abandoned, terminating")
		fsm.endGame(Terminated)
		return
	}

	if fsm.state == WaitingForConfirms {
		fsm.checkConfirms()
		return
	}
}

// TODO: return error
func (fsm *FSM) PlayerConfirmed(name string) {
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

func (fsm *FSM) checkConfirms() {
	if fsm.confirms.Size() == fsm.players.Size() {
		log.Logger.Info("Every player has confirmed, starting next round")
		fsm.resetConfirms()
		go fsm.runNextRound()
	}
}

func (fsm *FSM) runNextRound() {
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

func (fsm *FSM) countdown() bool {
	fsm.timer = time.NewTimer(fsm.getCountdownDuration())

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

func (fsm *FSM) playRound() bool {
	fsm.timer = time.NewTimer(fsm.getRoundDuration())

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

func (fsm *FSM) updateState(state FSMState) {
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

func (fsm *FSM) endGame(withState FSMState) {
	fsm.updateState(withState)
	close(fsm.cancelTime)
	close(fsm.updates)
}

func (fsm *FSM) isCountingDownToFirstRound() bool {
	return fsm.state == CountingDown && fsm.round == 1
}

func (fsm *FSM) isGameFull() bool {
	return fsm.players.Size() == fsm.config.MaxPlayers
}

func (fsm *FSM) isGameEmpty() bool {
	return fsm.players.Size() == 0
}

func (fsm *FSM) isGameAbandoned() bool {
	return fsm.isGameEmpty() && fsm.state != WaitingForJoins
}

func (fsm *FSM) hasGameEnded() bool {
	return fsm.state == Done || fsm.state == Terminated
}

func (fsm *FSM) isLastRound() bool {
	return fsm.round == fsm.config.Rounds
}

func (fsm *FSM) resetConfirms() {
	fsm.confirms = utils.NewSet[string]()
}

func (fsm *FSM) getRoundDuration() time.Duration {
	return time.Duration(fsm.config.RoundDuration) * time.Second
}

func (fsm *FSM) getCountdownDuration() time.Duration {
	return time.Duration(fsm.config.CountdownDuration) * time.Second
}
