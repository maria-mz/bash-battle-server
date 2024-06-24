package server

import (
	"io"

	"github.com/maria-mz/bash-battle-proto/proto"
)

type streamEndMsg struct {
	info string
	err  error
}

type stream struct {
	streamSrv     proto.BashBattle_StreamServer
	ackMsgs       chan *proto.AckMsg
	endStreamMsgs chan streamEndMsg
	done          bool
}

func NewStream(streamServer proto.BashBattle_StreamServer) *stream {
	return &stream{
		streamSrv:     streamServer,
		ackMsgs:       make(chan *proto.AckMsg),
		endStreamMsgs: make(chan streamEndMsg),
	}
}

func (s *stream) Recv() {
	if s.done {
		return
	}

	for {
		msg, err := s.streamSrv.Recv()

		if err == io.EOF { // happens when client calls CloseSend()
			s.closeStream(streamEndMsg{info: "EOF"})
			return
		}

		if err != nil {
			s.closeStream(streamEndMsg{err: err})
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
		s.closeStream(streamEndMsg{err: err})
	}
}

func (s *stream) closeStream(msg streamEndMsg) {
	s.endStreamMsgs <- msg
	close(s.endStreamMsgs)
	close(s.ackMsgs)
	s.done = true
}
