// Package store provides a thread-safe in-memory key-value store with TTL support.
package store

import (
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt *time.Time // nil means no expiry
}

// Store is a concurrent-safe key-value store inspired by Redis.
type Store struct {
	mu      sync.RWMutex
	data    map[string]entry
	stopCh  chan struct{}
	cleaned int // total expired keys evicted
}

// New creates and starts a store with a background TTL eviction loop.
func New() *Store {
	s := &Store{
		data:   make(map[string]entry),
		stopCh: make(chan struct{}),
	}
	go s.ttlLoop()
	return s
}

// Set stores a key-value pair with optional TTL (0 = no expiry).
func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}

	s.data[key] = entry{value: value, expiresAt: expiresAt}
}

// Get retrieves a value by key. Returns ("", false) if key not found or expired.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}

	if e.isExpired() {
		s.mu.Lock()
		delete(s.data, key)
		s.cleaned++
		s.mu.Unlock()
		return "", false
	}

	return e.value, true
}

// Del removes one or more keys. Returns the count of keys actually removed.
func (s *Store) Del(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, k := range keys {
		if _, ok := s.data[k]; ok {
			delete(s.data, k)
			count++
		}
	}
	return count
}

// Exists returns 1 if the key exists and is not expired, 0 otherwise.
func (s *Store) Exists(key string) int {
	_, ok := s.Get(key)
	if ok {
		return 1
	}
	return 0
}

// TTL returns the remaining TTL for a key in seconds, or -2 if not found, -1 if no expiry.
func (s *Store) TTL(key string) int64 {
	s.mu.RLock()
	e, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return -2
	}

	if e.isExpired() {
		s.mu.Lock()
		delete(s.data, key)
		s.cleaned++
		s.mu.Unlock()
		return -2
	}

	if e.expiresAt == nil {
		return -1
	}

	remaining := time.Until(*e.expiresAt).Seconds()
	if remaining < 0 {
		return -2
	}
	return int64(remaining)
}

// Expire sets a TTL on an existing key. Returns true if the key existed.
func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.data[key]
	if !ok {
		return false
	}

	t := time.Now().Add(ttl)
	e.expiresAt = &t
	s.data[key] = e
	return true
}

// Keys returns all non-expired keys matching an optional pattern.
// For now, we support only "*" (all keys).
func (s *Store) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Clean expired keys first
	now := time.Now()
	keys := make([]string, 0, len(s.data))
	for k, e := range s.data {
		if e.expiresAt != nil && now.After(*e.expiresAt) {
			delete(s.data, k)
			s.cleaned++
			continue
		}
		if pattern == "*" || pattern == "" {
			keys = append(keys, k)
		}
	}
	return keys
}

// Len returns the number of non-expired keys.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	count := 0
	for k, e := range s.data {
		if e.expiresAt != nil && now.After(*e.expiresAt) {
			delete(s.data, k)
			s.cleaned++
			continue
		}
		count++
	}
	return count
}

// Stats returns store statistics.
func (s *Store) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	live := 0
	expired := 0
	for _, e := range s.data {
		if e.expiresAt != nil && now.After(*e.expiresAt) {
			expired++
		} else {
			live++
		}
	}

	return map[string]interface{}{
		"keys":       live,
		"expired":    s.cleaned + expired,
		"max_memory": "unlimited",
	}
}

// Flush removes all keys.
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]entry)
}

// Shutdown stops the background TTL loop.
func (s *Store) Shutdown() {
	close(s.stopCh)
}

// --- Internal ---

func (e entry) isExpired() bool {
	if e.expiresAt == nil {
		return false
	}
	return time.Now().After(*e.expiresAt)
}

func (s *Store) ttlLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.evictExpired()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Store) evictExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, e := range s.data {
		if e.expiresAt != nil && now.After(*e.expiresAt) {
			delete(s.data, k)
			s.cleaned++
		}
	}
}