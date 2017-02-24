package types

import (
	"sync"
)

var pool = sync.Pool{}

// Set is an implementation of ISet using the builtin map type. Set is threadsafe.
type Set struct {
	items     map[string]bool
	lock      sync.RWMutex
	flattened []string
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

// Flatten will return a list of the items in the set.
func (set *Set) Flatten() []string {
	set.lock.Lock()
	defer set.lock.Unlock()

	if set.flattened != nil {
		return set.flattened
	}

	set.flattened = make([]string, 0, len(set.items))
	for item := range set.items {
		set.flattened = append(set.flattened, item)
	}
	return set.flattened
}

// Len returns the number of items in the set.
func (set *Set) Len() int64 {
	set.lock.RLock()
	size := int64(len(set.items))
	set.lock.RUnlock()

	return size
}

// Clear will remove all items from the set.
func (set *Set) Clear() {
	set.lock.Lock()
	set.items = map[string]bool{}
	set.lock.Unlock()
}

// New is the constructor for sets.  It will pull from a reuseable memory pool if it can.
// Takes a list of items to initialize the set with.
func NewSet(items ...string) *Set {
	set := pool.Get().(*Set)
	for _, item := range items {
		set.items[item] = true
	}

	return set
}

func init() {
	pool.New = func() interface{} {
		return &Set{
			items: make(map[string]bool, 10),
		}
	}
}
