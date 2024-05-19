package registry

type IdentifiableRecord[T any] interface {
	ID() T
}

type Registry[K comparable, V IdentifiableRecord[K]] struct {
	Records map[K]V
}

// Registry creates an empty Registry
func NewRegistry[K comparable, V IdentifiableRecord[K]]() *Registry[K, V] {
	return &Registry[K, V]{
		Records: make(map[K]V),
	}
}

// HasRecord checks if a record with the given id exists in the registry.
func (reg *Registry[K, V]) HasRecord(id K) bool {
	_, ok := reg.Records[id]
	return ok
}

// GetRecord returns a record matching the id, if it exists.
func (reg *Registry[K, V]) GetRecord(id K) (V, bool) {
	record, ok := reg.Records[id]
	return record, ok
}

// WriteRecord adds, or updates, a record in the registry.
func (reg *Registry[K, V]) WriteRecord(record V) {
	reg.Records[record.ID()] = record
}

// DeleteRecord deletes a record matching the id. If matching record is found,
// DeleteRecord is a no-op.
func (reg *Registry[K, V]) DeleteRecord(id K) {
	delete(reg.Records, id)
}
