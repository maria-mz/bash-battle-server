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
	err = s.Start()

	if err != nil {
		slog.Error("failed to start server")
		os.Exit(1)
	}

	slog.Info("server is running!")

	time.Sleep(10 * time.Second)

	s.Shutdown()
}
