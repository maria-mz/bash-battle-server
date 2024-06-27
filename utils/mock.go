package utils

import (
	"io"

	pb "github.com/maria-mz/bash-battle-proto/proto"
)

type MockStreamServer struct {
	AckMsgs        chan *pb.AckMsg
	RecievedEvents chan *pb.Event
	stop           chan bool
}

func NewMockStreamServer() *MockStreamServer {
	return &MockStreamServer{
		AckMsgs:        make(chan *pb.AckMsg, 10),
		RecievedEvents: make(chan *pb.Event, 10),
		stop:           make(chan bool),
	}
}

func (mss *MockStreamServer) Send(e *pb.Event) error {
	mss.RecievedEvents <- e
	return nil
}

func (mss *MockStreamServer) Recv() (*pb.AckMsg, error) {
	select {
	case msg := <-mss.AckMsgs:
		return msg, nil
	case <-mss.stop:
		return nil, io.EOF // The graceful error
	}
}

// (for testing, not part of simpleStreamServer interface)
func (mss *MockStreamServer) Close() {
	mss.stop <- true
}
