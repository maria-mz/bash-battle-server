package utils

// StrSet represents a set of strings.
type StrSet map[string]struct{}

// NewStrSet creates and returns a new empty StrSet.
func NewStrSet() StrSet {
	return make(StrSet)
}

// Add inserts an item into the set.
func (s StrSet) Add(item string) {
	s[item] = struct{}{}
}

// Delete removes an item from the set.
func (s StrSet) Delete(item string) {
	delete(s, item)
}

// Contains checks if the set contains the specified item.
func (s StrSet) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

// Size returns the number of items in the set.
func (s StrSet) Size() int {
	return len(s)
}

// Clear removes all items from the set.
func (s StrSet) Clear() {
	for k := range s {
		delete(s, k)
	}
}

// Items returns a slice containing all items in the set.
func (s StrSet) Items() []string {
	items := make([]string, 0, len(s))
	for k := range s {
		items = append(items, k)
	}
	return items
}
