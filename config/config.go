package config

import (
	"encoding/json"
	"os"
)

type GameConfig struct {
	MaxPlayers        int
	Rounds            int
	RoundDuration     int
	CountdownDuration int
	Difficulty        int
	FileSize          int
}

type Config struct {
	Host       string     `json:"host"`
	Port       uint16     `json:"port"`
	GameConfig GameConfig `json:"gameConfig"`
}

func LoadConfig() (Config, error) {
	var config Config

	file, err := os.Open("config/config.json")

	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)

	if err != nil {
		return config, err
	}

	return config, nil
}
