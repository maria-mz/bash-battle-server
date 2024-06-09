package data

import "github.com/maria-mz/bash-battle-server/config"

type FilePath string

type Challenge struct {
	Question   string
	InputFile  FilePath
	OutputFile FilePath
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
