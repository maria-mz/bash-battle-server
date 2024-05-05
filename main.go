package main

import (
	"log/slog"
	"os"
	"time"
)

// Spins up server and lets it run for 10s
// TODO: Run forever, close on ctrl+c or other signals
func main() {
	config, err := LoadConfig()

	if err != nil {
		slog.Error("failed to load server config", err)
		os.Exit(1)
	}

	s := NewServer(config)
	s.Start()

	time.Sleep(10 * time.Second)

	s.Shutdown()
}
