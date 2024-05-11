package main

import (
	"sync"
)

// Registry is a generic registry structure.
// It holds records by unique identifiers.
type Registry struct {
	mu      sync.RWMutex
	records map[ID]IdentifiableRecord
}

// NewRegistry creates a new instance of Registry
func NewRegistry() *Registry {
	return &Registry{
		records: make(map[ID]IdentifiableRecord),
	}
}

// GetRecord returns a record and a success flag
func (registry *Registry) GetRecord(id ID) (*IdentifiableRecord, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	record, ok := registry.records[id]
	return &record, ok
}

// SetRecord sets a record in the registry.
// If there is already a record with the same ID, it will be overwritten with
// the given record. Otherwise, it will add a new record.
// Returns a flag indicating whether a new record was created.
func (registry *Registry) SetRecord(record IdentifiableRecord) bool {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	_, ok := registry.records[record.ID()]

	registry.records[record.ID()] = record

	return !ok
}

// DeleteRecord deletes a record from the registry
func (registry *Registry) DeleteRecord(id ID) {
	delete(registry.records, id)
}
