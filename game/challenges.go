package game

import (
	"fmt"

	"github.com/maria-mz/bash-battle-server/config"
)

type FilePath string

type Challenge struct {
	Question   string
	InputFile  FilePath
	OutputFile FilePath
}

func (challenge *Challenge) InfoString() string {
	return fmt.Sprintf("%+v", challenge)
}

// TODO: temporary, implement real functionality, randomly assign rounds, use db
func GenerateChallenges(config config.GameConfig) map[int]Challenge {
	challenges := make(map[int]Challenge)

	tempChallenge := Challenge{
		Question:   "???",
		InputFile:  "input.txt",
		OutputFile: "output.txt",
	}

	for i := 0; i < config.Rounds; i++ {
		challenges[i] = tempChallenge
	}

	return challenges
}
