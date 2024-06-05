package server

import (
	"github.com/maria-mz/bash-battle-server/registry"
)

type ClientRegistry struct {
	clients *registry.Registry[string, Client]
}

func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		clients: registry.NewRegistry[string, Client](),
	}
}

func (r *ClientRegistry) AddClient(token string, username string) {
	client := &Client{
		token:    token,
		username: username,
	}
	r.clients.WriteRecord(token, client)
}

func (r *ClientRegistry) GetClient(token string) (*Client, bool) {
	client, ok := r.clients.GetRecord(token)
	return client, ok
}

func (r *ClientRegistry) HasClient(token string) bool {
	ok := r.clients.HasRecord(token)
	return ok
}
