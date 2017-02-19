package types

import (
	"sync"
)

// Set is an implementation of ISet using the builtin map type. Set is threadsafe.
type Set struct {
	items     map[string]bool
	lock      sync.RWMutex
}

// Add will add the provided items to the set.
func (set *Set) Add(item string) {
	set.lock.Lock()
	set.items[item] = true
	set.lock.Unlock()
}

// Exists returns a bool indicating if the given item exists in the set.
func (set *Set) Exists(item string) bool {
	set.lock.RLock()
	_, ok := set.items[item]
	set.lock.RUnlock()

	return ok
}

// New is the constructor for sets.  It will pull from a reuseable memory pool if it can.
// Takes a list of items to initialize the set with.
func New() *Set {
	return &Set{
		items: make(map[string]bool, 10),
	}
}