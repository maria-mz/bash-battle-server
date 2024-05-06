package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

var TestConfig = ServerConfig{
	Host: "127.0.0.1",
	Port: 3000, // Not the real port
}

var TestServer *Server

func TestMain(m *testing.M) {
	TestServer = NewServer(TestConfig)
	err := TestServer.Start()

	if err != nil {
		log.Println("failed to start server")
		os.Exit(1)
	}

	log.Println("server is running!")

	result := m.Run()

	log.Println("shutting down server")
	TestServer.Shutdown()

	os.Exit(result)
}

func testAcceptXConnections(t *testing.T, remoteAddr string, x int) {
	var w sync.WaitGroup

	localAddrs := make(chan string, x)

	for i := 0; i < x; i++ {
		w.Add(1)

		go func() {
			defer w.Done()

			conn, err := net.Dial("tcp", remoteAddr)

			if err != nil {
				log.Printf("failed to connect to server: %v", err)
				return
			}

			localAddrs <- conn.LocalAddr().String()
		}()
	}

	w.Wait()

	close(localAddrs)

	// Some leeway to make sure all have been processed
	time.Sleep(500 * time.Millisecond)

	if actual := len(TestServer.connections); actual < x {
		t.Fatalf(
			"expected %d connections in the map but got %d", x, actual,
		)
	}

	for addr := range localAddrs {
		if _, ok := TestServer.connections[addr]; !ok {
			t.Errorf("local address %s not in connections", addr)
		}
	}
}

func TestAcceptConnections(t *testing.T) {
	remoteAddress := fmt.Sprintf("%s:%d", TestConfig.Host, TestConfig.Port)

	// TODO: paramaterize, can't easily because server holds onto connections
	// need a way to close connections
	testAcceptXConnections(t, remoteAddress, 10)
}

// TODO: Test read
