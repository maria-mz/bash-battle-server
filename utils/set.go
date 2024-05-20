package utils

// Set represents a generic set.
type Set[T comparable] map[T]struct{}

// NewSet creates and returns a new empty Set.
func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

// Add inserts an item into the set.
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Delete removes an item from the set.
func (s Set[T]) Delete(item T) {
	delete(s, item)
}

// Contains checks if the set contains the specified item.
func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

// Size returns the number of items in the set.
func (s Set[T]) Size() int {
	return len(s)
}

// Clear removes all items from the set.
func (s Set[T]) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// Items returns a slice containing all items in the set.
func (s Set[T]) Items() []T {
	items := make([]T, 0, len(s))
	for k := range s {
		items = append(items, k)
	}
	return items
}
