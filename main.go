package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	reg "github.com/maria-mz/bash-battle-server/registry"
	srv "github.com/maria-mz/bash-battle-server/server"
	"google.golang.org/grpc"
)

var listener net.Listener
var server *srv.Server
var serverRegistrar *grpc.Server

func listen(host string, port uint16) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	listener = lis
}

func initServer() {
	games := reg.NewRegistry[string, srv.GameRecord]()
	clients := reg.NewRegistry[string, srv.ClientRecord]()
	server = srv.NewServer(clients, games)
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
		log.Printf("received signal: %v, shutting down server...", sig)

		serverRegistrar.GracefulStop()
		listener.Close()
		os.Exit(0)
	}()
}

func startServer() {
	err := serverRegistrar.Serve(listener)

	if err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func main() {
	config, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("failed to load server config: %s", err)
	}

	log.Printf(
		"configuring server on host: %s, port: %d",
		config.Host,
		config.Port,
	)

	listen(config.Host, config.Port)
	defer listener.Close()

	initServer()
	registerServer()

	handleSignals()

	log.Printf("starting server")
	startServer()
}
