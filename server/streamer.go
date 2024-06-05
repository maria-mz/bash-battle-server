package server

import (
	"fmt"

	"github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/log"
	"github.com/maria-mz/bash-battle-server/registry"
)

type Streamer struct {
	streams      *registry.Registry[string, Stream]
	incomingMsgs chan<- IncomingMsg
}

func NewStreamer(incMsgs chan<- IncomingMsg) *Streamer {
	return &Streamer{
		streams:      registry.NewRegistry[string, Stream](),
		incomingMsgs: incMsgs,
	}
}

func (s *Streamer) StartStreaming(sid string, streamServer proto.BashBattle_StreamServer) error {
	if s.streams.HasRecord(sid) {
		return fmt.Errorf("stream with sid %s is already active", sid)
	}

	stream := NewStream(sid, streamServer)
	s.streams.WriteRecord(sid, stream)

	go stream.Receive(s.incomingMsgs)

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

	s.streams.DeleteRecord(sid)

	return msg.err
}

func (s *Streamer) Broadcast(event *proto.Event, eventName string) {
	log.Logger.Debug("Broadcast event received", "event", event)
	log.Logger.Info("Broadcasting event", "event", eventName, "streams", s.streams.Size())

	for _, stream := range s.streams.Records() {
		stream.Send(event)
	}
}
