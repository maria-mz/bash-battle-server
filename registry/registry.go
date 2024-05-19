package registry

type IdentifiableRecord interface {
	ID() string
}

type Registry struct {
	Records map[string]IdentifiableRecord
}

// Registry creates an empty Registry
func NewRegistry() *Registry {
	return &Registry{
		Records: make(map[string]IdentifiableRecord),
	}
}

// HasRecord checks if a record with the given id exists in the registry.
func (reg *Registry) HasRecord(id string) bool {
	_, ok := reg.Records[id]
	return ok
}

// GetRecord returns a record matching the id, if it exists.
func (reg *Registry) GetRecord(id string) (IdentifiableRecord, bool) {
	record, ok := reg.Records[id]
	return record, ok
}

// WriteRecord adds, or updates, a record in the registry.
func (reg *Registry) WriteRecord(record IdentifiableRecord) {
	reg.Records[record.ID()] = record
}

// DeleteRecord deletes a record matching the id. If matching record is found,
// DeleteRecord is a no-op.
func (reg *Registry) DeleteRecord(id string) {
	delete(reg.Records, id)
}
