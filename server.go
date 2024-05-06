package main

import (
	"fmt"
	"log/slog"
	"net"
)

type Server struct {
	config      ServerConfig
	listener    net.Listener
	connections map[string]net.Conn
	closing     chan struct{}
	messageBus  *MessageBus
}

func NewServer(config ServerConfig) *Server {
	return &Server{
		config:      config,
		connections: make(map[string]net.Conn),
		closing:     make(chan struct{}),
		messageBus:  NewMessageBus(),
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	listener, err := net.Listen("tcp", address)

	if err != nil {
		return err
	}

	s.listener = listener

	go s.messageBus.HandleIncomingMessages()

	go s.acceptConnections()

	return nil
}

func (s *Server) acceptConnections() {
	defer s.listener.Close()

	for {
		select {
		case <-s.closing:
			slog.Info("stopped accepting connections")
			return
		default:
			conn, err := s.listener.Accept() // Blocking

			if err != nil {
				slog.Error(fmt.Sprintf("accept connection error: %v", err))
				continue
			}

			s.onAcceptConnection(conn)
		}
	}
}

func (s *Server) onAcceptConnection(conn net.Conn) {
	slog.Info(fmt.Sprintf("accepting new connection at %s", conn.RemoteAddr()))

	s.recordConnection(conn)
	go s.handleConnection(conn)
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	addr := conn.RemoteAddr().String()

	for {
		select {
		case <-s.closing:
			return
		default:
			n, err := conn.Read(buf) // Blocking

			if err != nil {
				slog.Error(fmt.Sprintf("read error for %s: %v", addr, err))
				delete(s.connections, addr)
				return
			}

			msg := Message{
				address: conn.RemoteAddr().String(),
				payload: buf[:n],
			}

			s.messageBus.PostMessage(msg)
		}
	}
}

func (server *Server) recordConnection(conn net.Conn) {
	server.connections[conn.RemoteAddr().String()] = conn
}

func (s *Server) Shutdown() {
	select {
	case <-s.closing:
		slog.Info("server is already shut down!")
	default:
		slog.Info("shutting down server")
		close(s.closing)
	}
}
