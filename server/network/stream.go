package network

import (
	"io"

	"github.com/maria-mz/bash-battle-proto/proto"
)

type SimpleStreamServer interface {
	Send(*proto.Event) error
	Recv() (*proto.AckMsg, error)
}

type EndStreamMsgs struct {
	Info string
	Err  error
}

type Stream struct {
	streamSrv SimpleStreamServer
	done      bool

	AckMsgs       chan *proto.AckMsg
	EndStreamMsgs chan EndStreamMsgs
}

func NewStream(streamSrv SimpleStreamServer) *Stream {
	return &Stream{
		streamSrv:     streamSrv,
		AckMsgs:       make(chan *proto.AckMsg),
		EndStreamMsgs: make(chan EndStreamMsgs),
	}
}

func (s *Stream) Recv() {
	if s.done {
		return
	}

	for {
		msg, err := s.streamSrv.Recv()

		if err == io.EOF { // happens when client calls CloseSend()
			s.closeStream(EndStreamMsgs{Info: "EOF"})
			return
		}

		if err != nil {
			s.closeStream(EndStreamMsgs{Err: err})
			return
		}

		s.AckMsgs <- msg
	}
}

func (s *Stream) SendEvent(event *proto.Event) {
	if s.done {
		return
	}

	if err := s.streamSrv.Send(event); err != nil {
		s.closeStream(EndStreamMsgs{Err: err})
	}
}

func (s *Stream) closeStream(msg EndStreamMsgs) {
	s.EndStreamMsgs <- msg
	close(s.EndStreamMsgs)
	close(s.AckMsgs)
	s.done = true
}
