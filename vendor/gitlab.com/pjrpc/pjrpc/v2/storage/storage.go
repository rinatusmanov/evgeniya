// Package storage contains storage of the internal handlers.
// Thread-safe put/get by string key.
package storage

import (
	"errors"
	"fmt"
	"sync"
)

// ErrRouteNotFound returns when storage can't found handler by route.
var ErrRouteNotFound = errors.New("route not found")

// Storage thread-safe map handlers by string routers.
type Storage struct {
	mu sync.Mutex

	storage map[string]interface{}
}

// New returns new storage.
func New() *Storage {
	s := &Storage{
		storage: make(map[string]interface{}),
	}

	return s
}

func (s *Storage) put(key string, value interface{}) {
	s.storage[key] = value
}

func (s *Storage) get(key string) (value interface{}, ok bool) {
	value, ok = s.storage[key]
	return
}

// Put puts new handler into storage.
func (s *Storage) Put(route string, value interface{}) {
	s.mu.Lock()
	s.put(route, value)
	s.mu.Unlock()
}

// Get returns handler from storage by route key.
// Returns error if route does not exist in storage.
func (s *Storage) Get(route string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.get(route)
	if !ok {
		return nil, fmt.Errorf("%w: '%s'", ErrRouteNotFound, route)
	}

	return value, nil
}
