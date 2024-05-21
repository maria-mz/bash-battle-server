package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/game"
	"github.com/maria-mz/bash-battle-server/log"
	srv "github.com/maria-mz/bash-battle-server/server"
	"google.golang.org/grpc"
)

var listener net.Listener
var server *srv.Server
var serverRegistrar *grpc.Server

func listen(host string, port uint16) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		log.Logger.Fatal("Failed to listen", "err", err)
	}

	listener = lis
}

func initServer() {
	// TODO: take in flags
	gameConfig := game.GameConfig{
		MaxPlayers:   4,
		Rounds:       10,
		RoundSeconds: 300,
		Difficulty:   game.VariedDiff,
		FileSize:     game.VariedSize,
	}

	clients := srv.NewRegistry[string, srv.ClientRecord]()

	server = srv.NewServer(clients, gameConfig)
}

func registerServer() {
	serverRegistrar = grpc.NewServer()
	proto.RegisterBashBattleServer(serverRegistrar, server)
}

func handleSignals() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-stop
		log.Logger.Info("Shutting down server", "signal", sig)

		serverRegistrar.GracefulStop()
		listener.Close()
		os.Exit(0)
	}()
}

func startServer() {
	err := serverRegistrar.Serve(listener)

	if err != nil {
		log.Logger.Fatal("Failed to serve", "err", err)
	}
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

	listen(config.Host, config.Port)
	defer listener.Close()

	initServer()
	registerServer()

	handleSignals()

	log.Logger.Info("Started server :)")
	startServer()
}
