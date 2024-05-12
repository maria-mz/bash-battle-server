package main

import (
	"fmt"
	"sync"
)

// -- Client Record

type ClientID string

// ClientRecord represents a client's record in the registry.
type ClientRecord struct {
	// ClientID uniquely identifies the client.
	ClientID ClientID

	// PlayerName is the name of the player associated with the client.
	PlayerName string

	// JoinedGameID is the ID of the game the client has joined, if any.
	JoinedGameID *GameID
}

// -- Client Registry Errors

type ErrClientRecordExists struct {
	ClientID ClientID
}

func (e ErrClientRecordExists) Error() string {
	return fmt.Sprintf("record with client ID %s already exists", e.ClientID)
}

type ErrPlayerNameTaken struct {
	PlayerName string
}

func (e ErrPlayerNameTaken) Error() string {
	return fmt.Sprintf("player name '%s' is already taken", e.PlayerName)
}

// ClientRegistry manages clients that log in to the server.
// Each client has a record, and this registry supports various operations
// on these records.
type ClientRegistry struct {
	// records maps ClientIDs to their corresponding ClientRecord.
	records map[ClientID]*ClientRecord

	// playerNamesSet is a set of player names currently in use.
	playerNamesSet map[string]struct{}

	// mu is used to make certain operations atomic.
	mu sync.RWMutex
}

// NewClientRegistry creates an empty ClientRegistry
func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		records:        make(map[ClientID]*ClientRecord),
		playerNamesSet: make(map[string]struct{}),
	}
}

// RegisterClient creates a new client record in the registry.
// If a record with the same client ID already exists, it returns an ErrClientRecordExists error.
// If the player name is already in use, it returns an ErrPlayerNameTaken error.
func (registry *ClientRegistry) RegisterClient(clientID ClientID, name string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if registry.HasRecord(clientID) {
		return ErrClientRecordExists{clientID}
	}

	if registry.isPlayerNameTaken(name) {
		return ErrPlayerNameTaken{name}
	}

	record := &ClientRecord{
		ClientID:   clientID,
		PlayerName: name,
	}

	registry.records[clientID] = record
	registry.addPlayerNameToSet(name)

	return nil
}

// isPlayerNameTaken checks if the given player name is already in use.
func (registry *ClientRegistry) isPlayerNameTaken(name string) bool {
	_, ok := registry.playerNamesSet[name]
	return ok
}

// addPlayerNameToSet adds a player name to the set of used player names.
func (registry *ClientRegistry) addPlayerNameToSet(name string) {
	registry.playerNamesSet[name] = struct{}{}
}

// HasRecord checks if there is a record for the given client ID.
func (registry *ClientRegistry) HasRecord(clientID ClientID) bool {
	_, ok := registry.records[clientID]
	return ok
}
