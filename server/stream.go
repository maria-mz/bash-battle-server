package server

import (
	"io"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/log"
)

type EndStreamMsg struct {
	info string
	err  error
}

type IncomingMsg struct {
	sid string
	msg *proto.AwkMsg
}

type Stream struct {
	sid       string
	streamSrv proto.BashBattle_StreamServer
	endStream chan EndStreamMsg
}

func NewStream(sid string, streamServer proto.BashBattle_StreamServer) *Stream {
	return &Stream{
		sid:       sid,
		streamSrv: streamServer,
		endStream: make(chan EndStreamMsg),
	}
}

func (s *Stream) Receive(msgs chan<- IncomingMsg) {
	for {
		msg, err := s.streamSrv.Recv()

		if err == io.EOF { // happens when client calls CloseSend()
			s.endStream <- EndStreamMsg{info: "EOF"}
			return
		}

		if err != nil {
			s.endStream <- EndStreamMsg{err: err}
			return
		}

		msgs <- IncomingMsg{sid: s.sid, msg: msg}
	}
}

func (s *Stream) Send(event *proto.Event) {
	err := s.streamSrv.Send(event)

	if err != nil {
		log.Logger.Warn("Failed to send event to stream", "sid", s.sid)

		s.endStream <- EndStreamMsg{err: err}

	} else {
		log.Logger.Info("Successfully sent event to stream", "sid", s.sid)
	}
}
