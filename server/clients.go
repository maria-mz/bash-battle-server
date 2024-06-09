package server

import (
	"errors"
	"sync"

	"github.com/maria-mz/bash-battle-server/utils"
)

var ErrUsernameTaken error = errors.New("someone with this username already exists")

type Clients struct {
	clients   map[string]*Client // username to client pairs
	usernames utils.Set[string]
	mu        sync.Mutex
}

func NewClients() *Clients {
	return &Clients{
		clients:   make(map[string]*Client),
		usernames: utils.NewSet[string](),
	}
}

func (clients *Clients) AddClient(token string, username string) error {
	clients.mu.Lock()
	defer clients.mu.Unlock()

	if clients.usernames.Contains(username) {
		return ErrUsernameTaken
	}

	client := &Client{
		token:    token,
		username: username,
	}

	clients.clients[token] = client
	clients.usernames.Add(username)

	return nil
}

func (clients *Clients) GetClient(token string) (*Client, bool) {
	client, ok := clients.clients[token]
	return client, ok
}

func (clients *Clients) HasClient(token string) bool {
	_, ok := clients.clients[token]
	return ok
}
