package store

import (
	"sync"
	"time"
)

type Value struct {
	Data   string
	Expiry *time.Time
}

// TODO: Extend Value to support multiple data types (hash, list, set, zset).
// Consider adding a Type tag and per-type fields, e.g.:
//   Type ValueType
//   Str  string
//   Hash map[string]string
//   List []string
//   ZSet *SortedSet
// Also add store methods for each type (HashSet/HashGet, ListPush, ZAdd, etc.)

type Store struct {
	mu   sync.RWMutex
	data map[string]Value
}

func New() *Store {
	return &Store{
		data: make(map[string]Value),
	}
}

func (s *Store) Set(key, value string, expireMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v := Value{Data: value}
	if expireMs > 0 {
		exp := time.Now().Add(time.Duration(expireMs) * time.Millisecond)
		v.Expiry = &exp
	}
	s.data[key] = v
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return "", false
	}

	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return "", false
	}

	return v.Data, true
}

func (s *Store) Delete(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, key := range keys {
		if _, exists := s.data[key]; exists {
			delete(s.data, key)
			count++
		}
	}
	return count
}

func (s *Store) Exists(keys ...string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, key := range keys {
		if v, ok := s.data[key]; ok {
			if v.Expiry == nil || now.Before(*v.Expiry) {
				count++
			}
		}
	}
	return count
}

func (s *Store) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0)
	now := time.Now()

	for k, v := range s.data {
		if v.Expiry != nil && now.After(*v.Expiry) {
			continue
		}
		// Simple pattern matching (* means all)
		if pattern == "*" || k == pattern {
			keys = append(keys, k)
		}
	}
	return keys
}

func (s *Store) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for k, v := range s.data {
		if v.Expiry != nil && now.After(*v.Expiry) {
			delete(s.data, k)
			count++
		}
	}
	return count
}

func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}
