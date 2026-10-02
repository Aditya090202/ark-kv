// Package store will provide concurrency-safe key-value storage.
package store

import "sync"

// Store will hold the node's in-memory key-value data.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

// TODO: Add concurrency-safe read, write, delete, and clear operations.

func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	return value, exists
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, existed := s.data[key]
	delete(s.data, key)
	return existed
}

func (s *Store) Clear_All() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.data {
		delete(s.data, key)
	}
}
