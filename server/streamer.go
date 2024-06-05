package server

import (
	"fmt"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/registry"
)

type Streamer struct {
	streams  *registry.Registry[string, Stream]
	recvMsgs chan<- StreamMsg
}

func NewStreamer(recvMsgs chan<- StreamMsg) *Streamer {
	return &Streamer{
		streams:  registry.NewRegistry[string, Stream](),
		recvMsgs: recvMsgs,
	}
}

func (s *Streamer) RegisterStream(sid string, streamServer proto.BashBattle_StreamServer) {
	stream := NewStream(sid, streamServer)
	s.streams.WriteRecord(sid, stream)
}

func (s *Streamer) HasStream(sid string) bool {
	return s.streams.HasRecord(sid)
}

func (s *Streamer) IsStreamActive(sid string) bool {
	stream, ok := s.streams.GetRecord(sid)

	if !ok {
		return false
	}

	return stream.isActive
}

func (s *Streamer) UnRegisterStream(sid string) {
	// TODO: Need a way to stop recv stream loop ???
	s.streams.DeleteRecord(sid)
}

func (s *Streamer) StartStreaming(sid string) error {
	stream, ok := s.streams.GetRecord(sid)

	if !ok {
		return fmt.Errorf("no stream matching sid %s", sid)
	}

	stream.isActive = true

	go stream.Receive(s.recvMsgs)

	msg := <-stream.endStream // blocking

	if msg.err != nil {
		log.Logger.Warn(
			"Stream ended due to error", "sid", stream.sid, "err", msg.err,
		)
	} else {
		log.Logger.Info(
			"Stream ended gracefully", "sid", stream.sid, "info", msg.info,
		)
	}

	stream.isActive = false

	return msg.err
}

func (s *Streamer) Broadcast(event *proto.Event, eventName string) {
	log.Logger.Debug("Broadcast event received", "event", event)
	log.Logger.Info("Broadcasting event", "event", eventName, "streams", s.streams.Size())

	for _, stream := range s.streams.Records() {
		stream.Send(event)
	}
}
