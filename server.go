package main

import (
	"fmt"
	"log/slog"
	"net"
	"time"
)

type Server struct {
	config      ServerConfig
	listener    net.Listener
	connections []*Connection // TODO: Map ?
	accepting   bool
}

func NewServer(config ServerConfig) *Server {
	return &Server{config: config}
}

func (server *Server) Start() error {
	address := fmt.Sprintf("%s:%d", server.config.Host, server.config.Port)

	listener, err := net.Listen("tcp", address)

	if err != nil {
		slog.Error("failed to start server", err)
		return err
	}

	slog.Info("server is running!")

	server.listener = listener

	go server.acceptConnections()

	return nil
}

func (server *Server) acceptConnections() {
	server.accepting = true

	for server.accepting {
		netConn, err := server.listener.Accept() // Blocking

		if err != nil {
			slog.Error("accept connection error", "err", err)
			continue
		}

		conn := NewConnection(netConn)

		server.recordConnection(conn)

		slog.Info("accepted new connection", "address", conn.address)

		go conn.Read()
	}

	slog.Debug("stopped accepting connections")
}

func (server *Server) recordConnection(conn *Connection) {
	server.connections = append(server.connections, conn)
}

func (server *Server) Shutdown() {
	slog.Info("shutting down server [start]")

	server.StopAccepting()
	server.listener.Close()

	// Sleep for a bit to make sure accept loop exists, so no new
	// connections are recorded while we are closing.
	time.Sleep(1 * time.Second)

	for _, conn := range server.connections {
		slog.Info("closing connection", "address", conn.address)

		conn.StopReading()
		conn.Close()
	}

	slog.Info("shutting down server [done]")
}

func (server *Server) StopAccepting() {
	if !server.accepting {
		slog.Debug("server.accepting is already false")
	}
	server.accepting = false
}
