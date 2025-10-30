package command

import (
	"fmt"
	"path/filepath"
	"strconv"

	"redis-from-scratch/internal/store"
)

// Pattern matching for KEYS command
type PatternMatcher struct {
	pattern string
}

func NewPatternMatcher(pattern string) *PatternMatcher {
	return &PatternMatcher{pattern: pattern}
}

// Match checks if a key matches the pattern using glob-style matching
// Supports: * (any chars), ? (single char), [abc] (char class), [^abc] (negated class)
func (pm *PatternMatcher) Match(key string) bool {
	ok, err := filepath.Match(pm.pattern, key)
	if err != nil {
		// Invalid pattern, no matches
		return false
	}
	return ok
}

// Updated KeysHandler with pattern support
type KeysHandler struct{}

func (h *KeysHandler) Execute(s *store.Store, args []string) Response {
	pattern := "*"
	if len(args) > 0 {
		pattern = args[0]
	}
	keys := s.Keys(pattern)
	return Response{Type: TypeArray, Value: keys}
}

// DEL handler
type DelHandler struct{}

func (h *DelHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'del' command")}
	}
	n := s.Delete(args...)
	return Response{Type: TypeInteger, Value: n}
}

// EXISTS handler
type ExistsHandler struct{}

func (h *ExistsHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'exists' command")}
	}
	n := s.Exists(args...)
	return Response{Type: TypeInteger, Value: n}
}

// SCAN handler - implements cursor-based iteration
type ScanHandler struct{}

func (h *ScanHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'scan' command")}
	}

	cursor, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid cursor")}
	}

	pattern := "*"
	count := int64(10) // default count

	// Parse options: MATCH pattern, COUNT count
	i := 1
	for i < len(args) {
		switch args[i] {
		case "MATCH":
			if i+1 < len(args) {
				pattern = args[i+1]
				i += 2
			} else {
				return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
			}
		case "COUNT":
			if i+1 < len(args) {
				c, err := strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid count")}
				}
				count = c
				i += 2
			} else {
				return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
			}
		default:
			return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
		}
	}

	nextCursor, keys, err := s.Scan(cursor, pattern, count)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}

	// Response format: [nextCursor, [keys...]] - nested array
	return Response{
		Type: TypeNestedArray,
		Value: map[string]interface{}{
			"cursor": fmt.Sprintf("%d", nextCursor),
			"keys":   keys,
		},
	}
}

// HSCAN handler for scanning hash fields
type HScanHandler struct{}

func (h *HScanHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'hscan' command")}
	}

	key := args[0]
	cursor, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid cursor")}
	}

	pattern := "*"
	count := int64(10)

	i := 2
	for i < len(args) {
		switch args[i] {
		case "MATCH":
			if i+1 < len(args) {
				pattern = args[i+1]
				i += 2
			} else {
				return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
			}
		case "COUNT":
			if i+1 < len(args) {
				c, err := strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid count")}
				}
				count = c
				i += 2
			} else {
				return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
			}
		default:
			return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
		}
	}

	nextCursor, fields, err := s.HashScan(key, cursor, pattern, count)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}

	// Response format: [nextCursor, [fields...]] - nested array
	return Response{
		Type: TypeNestedArray,
		Value: map[string]interface{}{
			"cursor": fmt.Sprintf("%d", nextCursor),
			"keys":   fields,
		},
	}
}

// Register SCAN handlers
// Add to handlers map in command.go:
// "SCAN":  &ScanHandler{},
// "HSCAN": &HScanHandler{},
