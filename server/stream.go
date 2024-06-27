package server

import (
	"io"

	"github.com/maria-mz/bash-battle-proto/proto"
)

type simpleStreamServer interface {
	Send(*proto.Event) error
	Recv() (*proto.AckMsg, error)
}

type endStreamMsgs struct {
	info string
	err  error
}

type stream struct {
	streamSrv     simpleStreamServer
	ackMsgs       chan *proto.AckMsg
	endStreamMsgs chan endStreamMsgs
	done          bool
}

func NewStream(streamSrv simpleStreamServer) *stream {
	return &stream{
		streamSrv:     streamSrv,
		ackMsgs:       make(chan *proto.AckMsg),
		endStreamMsgs: make(chan endStreamMsgs),
	}
}

func (s *stream) Recv() {
	if s.done {
		return
	}

	for {
		msg, err := s.streamSrv.Recv()

		if err == io.EOF { // happens when client calls CloseSend()
			s.closeStream(endStreamMsgs{info: "EOF"})
			return
		}

		if err != nil {
			s.closeStream(endStreamMsgs{err: err})
			return
		}

		s.ackMsgs <- msg
	}
}

func (s *stream) SendEvent(event *proto.Event) {
	if s.done {
		return
	}

	if err := s.streamSrv.Send(event); err != nil {
		s.closeStream(endStreamMsgs{err: err})
	}
}

func (s *stream) closeStream(msg endStreamMsgs) {
	s.endStreamMsgs <- msg
	close(s.endStreamMsgs)
	close(s.ackMsgs)
	s.done = true
}
