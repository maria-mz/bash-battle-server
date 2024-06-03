package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/service"
)

func handleSignals(service *service.Service) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-stop
		log.Logger.Info("Shutting down server gracefully", "signal", sig)

		service.Shutdown()
		os.Exit(0)
	}()
}

func main() {
	log.InitLogger()

	config, err := config.LoadConfig()

	if err != nil {
		log.Logger.Fatal("Failed to load server config", "err", err)
	}

	log.Logger.Info(
		"Configuring server", "host", config.Host, "port", config.Port,
	)

	s := service.NewService(config)

	go handleSignals(s)

	log.Logger.Info("Started server :)")
	err = s.Run()
	defer s.Shutdown()

	if err != nil {
		log.Logger.Fatal("Failed to serve", "err", err)
	}
}
