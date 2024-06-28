package config

import (
	"encoding/json"
	"os"

	"github.com/maria-mz/bash-battle-proto/proto"
)

type GameConfig struct {
	MaxPlayers        int
	Rounds            int
	RoundDuration     int
	CountdownDuration int
	Difficulty        int
	FileSize          int
}

func (config *GameConfig) ToProto() *proto.GameConfig {
	return &proto.GameConfig{
		MaxPlayers:   int32(config.MaxPlayers),
		Rounds:       int32(config.Rounds),
		RoundSeconds: int32(config.RoundDuration),
		Difficulty:   proto.Difficulty(config.Difficulty),
		FileSize:     proto.FileSize(config.FileSize),
	}
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
