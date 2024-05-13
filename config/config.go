package config

import (
	"encoding/json"
	"os"
)

type ServerConfig struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

func LoadConfig() (ServerConfig, error) {
	var config ServerConfig

	file, err := os.Open("config.json")

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
