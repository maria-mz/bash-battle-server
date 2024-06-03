package service

import (
	"errors"
	"fmt"
	"net"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/config"
	"github.com/maria-mz/bash-battle-server/server"
	"google.golang.org/grpc"
)

type Service struct {
	config          config.Config
	listener        net.Listener
	server          *server.Server
	serverRegistrar *grpc.Server
}

func NewService(conf config.Config) *Service {
	s := &Service{}
	s.config = conf

	clients := server.NewRegistry[string, server.ClientRecord]()
	s.server = server.NewServer(clients, s.config.GameConfig)
	s.serverRegistrar = grpc.NewServer()

	proto.RegisterBashBattleServer(s.serverRegistrar, s.server)

	return s
}

func (s *Service) Run() error {
	if s.listener != nil {
		return errors.New("server is already running")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))

	if err != nil {
		return err
	}

	s.listener = lis

	err = s.serverRegistrar.Serve(s.listener) // blocking
	return err
}

func (s *Service) Shutdown() {
	// TODO: reverse order, see how ongoing streams are handled. do they hang?
	if s.serverRegistrar != nil {
		s.serverRegistrar.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
}
