package server

import (
	"fmt"
)

type ErrNameTaken struct {
	Name string
}

func (err ErrNameTaken) Error() string {
	return fmt.Sprintf("username '%s' is already taken", err.Name)
}

type ErrGameStarted struct{}

func (err ErrGameStarted) Error() string {
	return "the game has already started"
}

type ErrGameFull struct{}

func (err ErrGameFull) Error() string {
	return "the game is full"
}
