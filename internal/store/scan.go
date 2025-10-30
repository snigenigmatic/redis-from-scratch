// FILE: internal/store/scan.go
// PURPOSE: Cursor-based iteration methods for all data types
// Implements SCAN protocol with pattern matching support
// CREATE THIS AS A NEW FILE

package store

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"
)

// KeysPattern returns keys matching the given pattern
// Supports glob patterns: *, ?, [abc], [^abc]
func (s *Store) KeysPattern(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0)
	now := time.Now()

	for k, v := range s.data {
		// Skip expired keys
		if v.Expiry != nil && now.After(*v.Expiry) {
			continue
		}

		// Match against pattern
		ok, err := filepath.Match(pattern, k)
		if err != nil || !ok {
			continue
		}

		keys = append(keys, k)
	}

	// Sort for consistent output
	sort.Strings(keys)
	return keys
}

// Scan implements cursor-based iteration over keys
// Returns: nextCursor, keys, error
// cursor=0 starts from beginning; when nextCursor=0, iteration is complete
func (s *Store) Scan(cursor int64, pattern string, count int64) (int64, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if count <= 0 {
		count = 10
	}

	// Get all valid keys (not expired)
	allKeys := make([]string, 0)
	now := time.Now()

	for k, v := range s.data {
		if v.Expiry != nil && now.After(*v.Expiry) {
			continue
		}

		// Check if matches pattern
		ok, err := filepath.Match(pattern, k)
		if err != nil || !ok {
			continue
		}

		allKeys = append(allKeys, k)
	}

	// Sort for consistent iteration
	sort.Strings(allKeys)

	// Validate cursor
	if cursor < 0 {
		return 0, nil, fmt.Errorf("ERR invalid cursor")
	}

	// Determine slice bounds
	start := cursor
	end := cursor + count

	if start >= int64(len(allKeys)) {
		// Cursor beyond range, iteration complete
		return 0, []string{}, nil
	}

	if end > int64(len(allKeys)) {
		end = int64(len(allKeys))
	}

	result := allKeys[start:end]
	nextCursor := int64(0)

	if end < int64(len(allKeys)) {
		nextCursor = end
	}

	return nextCursor, result, nil
}

// HashScan implements cursor-based iteration over hash fields
func (s *Store) HashScan(key string, cursor int64, pattern string, count int64) (int64, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0, []string{}, nil
	}

	if v.Type != TypeHash {
		return 0, nil, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}

	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return 0, []string{}, nil
	}

	if count <= 0 {
		count = 10
	}

	// Get all matching fields
	allFields := make([]string, 0)
	for f := range v.Hash {
		ok, err := filepath.Match(pattern, f)
		if err != nil || !ok {
			continue
		}
		allFields = append(allFields, f)
	}

	// Sort for consistency
	sort.Strings(allFields)

	// Validate cursor
	if cursor < 0 {
		return 0, nil, fmt.Errorf("ERR invalid cursor")
	}

	// Determine slice bounds
	start := cursor
	end := cursor + count

	if start >= int64(len(allFields)) {
		return 0, []string{}, nil
	}

	if end > int64(len(allFields)) {
		end = int64(len(allFields))
	}

	// Build result as field, value, field, value...
	result := make([]string, 0)
	for _, f := range allFields[start:end] {
		result = append(result, f, v.Hash[f])
	}

	nextCursor := int64(0)
	if end < int64(len(allFields)) {
		nextCursor = end
	}

	return nextCursor, result, nil
}

// SetScan implements cursor-based iteration over set members
func (s *Store) SetScan(key string, cursor int64, pattern string, count int64) (int64, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0, []string{}, nil
	}

	if v.Type != TypeSet {
		return 0, nil, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}

	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return 0, []string{}, nil
	}

	if count <= 0 {
		count = 10
	}

	// Get all matching members
	allMembers := make([]string, 0)
	for m := range v.Set {
		ok, err := filepath.Match(pattern, m)
		if err != nil || !ok {
			continue
		}
		allMembers = append(allMembers, m)
	}

	// Sort for consistency
	sort.Strings(allMembers)

	// Validate cursor
	if cursor < 0 {
		return 0, nil, fmt.Errorf("ERR invalid cursor")
	}

	// Determine slice bounds
	start := cursor
	end := cursor + count

	if start >= int64(len(allMembers)) {
		return 0, []string{}, nil
	}

	if end > int64(len(allMembers)) {
		end = int64(len(allMembers))
	}

	result := allMembers[start:end]
	nextCursor := int64(0)

	if end < int64(len(allMembers)) {
		nextCursor = end
	}

	return nextCursor, result, nil
}
