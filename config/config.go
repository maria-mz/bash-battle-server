package config

import (
	"encoding/json"
	"os"

	"github.com/maria-mz/bash-battle-proto/proto"
)

type Config struct {
	Host       string            `json:"host"`
	Port       uint16            `json:"port"`
	GameConfig *proto.GameConfig `json:"gameConfig"`
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
