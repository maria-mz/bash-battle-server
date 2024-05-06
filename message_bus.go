package main

import "fmt"

type Message struct {
	address string
	payload []byte
}

type MessageBus struct {
	incomingMsgs chan Message
	outgoingMsgs chan Message
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		incomingMsgs: make(chan Message),
		outgoingMsgs: make(chan Message),
	}
}

func (bus *MessageBus) PostMessage(msg Message) {
	bus.incomingMsgs <- msg
}

func (bus *MessageBus) HandleIncomingMessages() {
	for msg := range bus.incomingMsgs {
		fmt.Printf("received message! %s\n", string(msg.payload))
	}
}
