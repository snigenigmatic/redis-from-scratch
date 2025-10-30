package store

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Value struct {
	// Type indicates the data type stored at this key. Start with TypeString = 0
	// to preserve backward compatibility with existing code that treated Value
	// as a string.
	Type ValueType

	// Str holds plain string values (SET/GET).
	Str string

	// Hash, List, Set and ZSet are placeholders for future data types.
	// Only one of these should be used depending on Type.
	Hash map[string]string
	List []string
	Set  map[string]struct{}
	ZSet *SortedSet

	Expiry *time.Time
}

// ValueType represents the stored value's data type.
type ValueType int

const (
	TypeString ValueType = iota
	TypeHash
	TypeList
	TypeSet
	TypeZSet
)

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

	v := Value{Type: TypeString, Str: value}
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

	if v.Type != TypeString {
		// Not a plain string value
		return "", false
	}

	return v.Str, true
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

// HashSet sets the field in the hash stored at key. Returns 1 if field is new, 0 if updated.
// Returns an error if the key exists and is not a hash.
func (s *Store) HashSet(key, field, value string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok && v.Type != TypeHash {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if !ok {
		v = Value{Type: TypeHash, Hash: make(map[string]string)}
	}
	_, existed := v.Hash[field]
	v.Hash[field] = value
	s.data[key] = v
	if existed {
		return 0, nil
	}
	return 1, nil
}

// HashGet returns the value associated with field in the hash stored at key.
// Returns ("", false, nil) if key or field does not exist. Returns an error if the key exists
// and is not a hash.
func (s *Store) HashGet(key, field string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return "", false, nil
	}
	if v.Type != TypeHash {
		return "", false, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	val, ok := v.Hash[field]
	if !ok {
		return "", false, nil
	}
	// Check expiry
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return "", false, nil
	}
	return val, true, nil
}

// HashDel deletes fields from the hash stored at key. Returns the number of fields removed.
// Returns an error if the key exists and is not a hash.
func (s *Store) HashDel(key string, fields ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return 0, nil
	}
	if v.Type != TypeHash {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	count := 0
	for _, f := range fields {
		if _, exists := v.Hash[f]; exists {
			delete(v.Hash, f)
			count++
		}
	}
	// If hash becomes empty, you could delete the key entirely
	if len(v.Hash) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = v
	}
	return count, nil
}

// HashGetAll returns a copy of the hash map at key. Returns an error if the key exists and is not a hash.
func (s *Store) HashGetAll(key string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return map[string]string{}, nil
	}
	if v.Type != TypeHash {
		return nil, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	// Copy to avoid exposing internal map
	out := make(map[string]string, len(v.Hash))
	for k, val := range v.Hash {
		out[k] = val
	}
	return out, nil
}

// ListLPush pushes values to the left of the list stored at key. Returns the new length.
// If the key does not exist, create a new list. Returns an error if key exists and is not a list.
func (s *Store) ListLPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok {
		// If expired, treat as not exist
		if v.Expiry != nil && time.Now().After(*v.Expiry) {
			delete(s.data, key)
			ok = false
		}
	}
	if ok && v.Type != TypeList {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if !ok {
		v = Value{Type: TypeList, List: make([]string, 0)}
	}
	// Prepend values in order: LPUSH a b c -> pushes a then b then c => list becomes c b a
	for i := 0; i < len(values); i++ {
		v.List = append([]string{values[i]}, v.List...)
	}
	s.data[key] = v
	return len(v.List), nil
}

// ListRPush pushes values to the right of the list stored at key. Returns the new length.
func (s *Store) ListRPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok {
		if v.Expiry != nil && time.Now().After(*v.Expiry) {
			delete(s.data, key)
			ok = false
		}
	}
	if ok && v.Type != TypeList {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if !ok {
		v = Value{Type: TypeList, List: make([]string, 0)}
	}
	v.List = append(v.List, values...)
	s.data[key] = v
	return len(v.List), nil
}

// ListLPop removes and returns the first element of the list stored at key.
// Returns ("", false, nil) if key does not exist or list is empty.
func (s *Store) ListLPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return "", false, nil
	}
	if v.Type != TypeList {
		return "", false, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		delete(s.data, key)
		return "", false, nil
	}
	if len(v.List) == 0 {
		return "", false, nil
	}
	val := v.List[0]
	v.List = v.List[1:]
	if len(v.List) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = v
	}
	return val, true, nil
}

// ListRPop removes and returns the last element of the list stored at key.
func (s *Store) ListRPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return "", false, nil
	}
	if v.Type != TypeList {
		return "", false, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		delete(s.data, key)
		return "", false, nil
	}
	if len(v.List) == 0 {
		return "", false, nil
	}
	last := v.List[len(v.List)-1]
	v.List = v.List[:len(v.List)-1]
	if len(v.List) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = v
	}
	return last, true, nil
}

// ListRange returns the elements between start and stop (inclusive).
// Supports negative indices like Redis (-1 is last element).
func (s *Store) ListRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}
	if v.Type != TypeList {
		return nil, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return []string{}, nil
	}
	ln := len(v.List)
	if ln == 0 {
		return []string{}, nil
	}
	// handle negative indices
	if start < 0 {
		start = ln + start
	}
	if stop < 0 {
		stop = ln + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= ln {
		stop = ln - 1
	}
	if start > stop || start >= ln {
		return []string{}, nil
	}
	return append([]string{}, v.List[start:stop+1]...), nil
}

// SetAdd adds the specified members to the set stored at key.
// Returns the number of elements that were added to the set (not including existing members).
// Returns an error if the key exists and is not a set.
func (s *Store) SetAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok {
		if v.Expiry != nil && time.Now().After(*v.Expiry) {
			delete(s.data, key)
			ok = false
		}
	}
	if ok && v.Type != TypeSet {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if !ok {
		v = Value{Type: TypeSet, Set: make(map[string]struct{})}
	}
	added := 0
	for _, m := range members {
		if _, exists := v.Set[m]; !exists {
			v.Set[m] = struct{}{}
			added++
		}
	}
	s.data[key] = v
	return added, nil
}

// SetMembers returns all the members of the set value stored at key.
// Returns an error if the key exists and is not a set.
func (s *Store) SetMembers(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}
	if v.Type != TypeSet {
		return nil, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return []string{}, nil
	}
	out := make([]string, 0, len(v.Set))
	for m := range v.Set {
		out = append(out, m)
	}
	return out, nil
}

// SetRemove removes the specified members from the set stored at key.
// Returns the number of members that were removed.
func (s *Store) SetRemove(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return 0, nil
	}
	if v.Type != TypeSet {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	removed := 0
	for _, m := range members {
		if _, exists := v.Set[m]; exists {
			delete(v.Set, m)
			removed++
		}
	}
	if len(v.Set) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = v
	}
	return removed, nil
}

// SetIsMember returns whether member is a member of the set stored at key.
// Returns (false, nil) if key does not exist. Returns an error if the key exists and is not a set.
func (s *Store) SetIsMember(key, member string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return false, nil
	}
	if v.Type != TypeSet {
		return false, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return false, nil
	}
	_, exists := v.Set[member]
	return exists, nil
}

// Sorted set implementation (simple slice + map). Not optimized for large sets.
type zEntry struct {
	member string
	score  float64
}

type SortedSet struct {
	entries []zEntry
	index   map[string]float64
}

func newSortedSet() *SortedSet {
	return &SortedSet{entries: make([]zEntry, 0), index: make(map[string]float64)}
}

// helper to find insertion index by score then member
func (ss *SortedSet) insertEntry(e zEntry) {
	i := sort.Search(len(ss.entries), func(i int) bool {
		if ss.entries[i].score == e.score {
			return ss.entries[i].member >= e.member
		}
		return ss.entries[i].score >= e.score
	})
	ss.entries = append(ss.entries, zEntry{})
	copy(ss.entries[i+1:], ss.entries[i:])
	ss.entries[i] = e
	ss.index[e.member] = e.score
}

func (ss *SortedSet) removeMember(member string) bool {
	score, ok := ss.index[member]
	if !ok {
		return false
	}
	idx := -1
	for i, e := range ss.entries {
		if e.member == member && e.score == score {
			idx = i
			break
		}
	}
	if idx == -1 {
		delete(ss.index, member)
		return false
	}
	ss.entries = append(ss.entries[:idx], ss.entries[idx+1:]...)
	delete(ss.index, member)
	return true
}

func (ss *SortedSet) getRange(start, stop int) []string {
	ln := len(ss.entries)
	if ln == 0 {
		return []string{}
	}
	if start < 0 {
		start = ln + start
	}
	if stop < 0 {
		stop = ln + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= ln {
		stop = ln - 1
	}
	if start > stop || start >= ln {
		return []string{}
	}
	out := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		out = append(out, ss.entries[i].member)
	}
	return out
}

// ZAdd: add member with score, return 1 if added, 0 if updated
func (s *Store) ZAdd(key string, score float64, member string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok {
		if v.Expiry != nil && time.Now().After(*v.Expiry) {
			delete(s.data, key)
			ok = false
		}
	}
	if ok && v.Type != TypeZSet {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if !ok {
		v = Value{Type: TypeZSet, ZSet: newSortedSet()}
	}
	ss := v.ZSet
	if old, exists := ss.index[member]; exists {
		if old == score {
			return 0, nil
		}
		ss.removeMember(member)
	}
	ss.insertEntry(zEntry{member: member, score: score})
	s.data[key] = v
	return 1, nil
}

// ZScore returns the score of member in the sorted set at key.
func (s *Store) ZScore(key, member string) (float64, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return 0, false, nil
	}
	if v.Type != TypeZSet {
		return 0, false, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return 0, false, nil
	}
	sc, exists := v.ZSet.index[member]
	return sc, exists, nil
}

// ZRange returns members in [start, stop]
func (s *Store) ZRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}
	if v.Type != TypeZSet {
		return nil, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	if v.Expiry != nil && time.Now().After(*v.Expiry) {
		return []string{}, nil
	}
	return v.ZSet.getRange(start, stop), nil
}

// ZRem removes members from the sorted set. Returns number removed.
func (s *Store) ZRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if !ok {
		return 0, nil
	}
	if v.Type != TypeZSet {
		return 0, fmt.Errorf("WRONGTYPE operation against a key holding the wrong kind of value")
	}
	removed := 0
	for _, m := range members {
		if v.ZSet.removeMember(m) {
			removed++
		}
	}
	if len(v.ZSet.entries) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = v
	}
	return removed, nil
}
