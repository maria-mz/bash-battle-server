package main

import (
	"fmt"
	"log/slog"
	"net"
)

type Connection struct {
	netConn net.Conn
	address string
	reading bool
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		netConn: conn,
		address: conn.RemoteAddr().String(),
	}
}

func (conn *Connection) Read() {
	conn.reading = true

	buf := make([]byte, 1024)

	for conn.reading {
		n, err := conn.netConn.Read(buf) // Blocking

		// TODO: count failed reads, if exceeds, send msg that
		// this connection seems lost, send msg to client to check, if
		// no response, close

		if err != nil {
			slog.Error("read error", "address", conn.address, "err", err)

			continue
		}

		msg := buf[:n]
		fmt.Println(string(msg))
	}

	slog.Debug("stopped reading", "address", conn.address)
}

func (conn *Connection) StopReading() {
	if !conn.reading {
		slog.Debug("conn.reading is already false", "address", conn.address)
		return
	}
	conn.reading = false
}

func (conn *Connection) Close() error {
	return conn.netConn.Close()
}
