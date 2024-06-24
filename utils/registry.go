package utils

type IdentifiableRecord[T any] interface{}

type Registry[K comparable, V IdentifiableRecord[K]] struct {
	records map[K]*V
}

// Registry creates an empty Registry
func NewRegistry[K comparable, V IdentifiableRecord[K]]() *Registry[K, V] {
	return &Registry[K, V]{
		records: make(map[K]*V),
	}
}

// HasRecord checks if a record with the given id exists in the registry.
func (reg *Registry[K, V]) HasRecord(id K) bool {
	_, ok := reg.records[id]
	return ok
}

// GetRecord returns a record matching the id, if it exists.
func (reg *Registry[K, V]) GetRecord(id K) (*V, bool) {
	record, ok := reg.records[id]
	return record, ok
}

// WriteRecord adds a record to the registry (or updates if not careful).
func (reg *Registry[K, V]) WriteRecord(id K, record *V) {
	reg.records[id] = record
}

// DeleteRecord deletes a record matching the id. If no matching record is found,
// DeleteRecord is a no-op.
func (reg *Registry[K, V]) DeleteRecord(id K) {
	delete(reg.records, id)
}

// Size returns the number of records in the registry.
func (reg *Registry[K, V]) Size() int {
	return len(reg.records)
}

// Records returns a slice of all the records in the registry.
func (reg *Registry[K, V]) Records() []*V {
	values := make([]*V, reg.Size())

	for _, v := range reg.records {
		values = append(values, v)
	}

	return values
}

// RecordsMatchingQuery returns the count of records satisfying the query.
func (reg *Registry[K, V]) RecordsMatchingQuery(query func(*V) bool) int {
	count := 0

	for _, v := range reg.records {
		if query(v) {
			count++
		}
	}

	return count
}
