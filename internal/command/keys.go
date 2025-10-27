package command

import (
	"fmt"

	"redis-from-scratch/internal/store"
)

type DelHandler struct{}

func (h *DelHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'del' command")}
	}
	count := s.Delete(args...)
	return Response{Type: TypeInteger, Value: count}
}

type ExistsHandler struct{}

func (h *ExistsHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'exists' command")}
	}
	count := s.Exists(args...)
	return Response{Type: TypeInteger, Value: count}
}

type KeysHandler struct{}

func (h *KeysHandler) Execute(s *store.Store, args []string) Response {
	pattern := "*"
	if len(args) > 0 {
		pattern = args[0]
	}
	keys := s.Keys(pattern)
	return Response{Type: TypeArray, Value: keys}
}

// TODO: Improve pattern matching for KEYS (support glob patterns like ? and []).
// Consider implementing SCAN for incremental iteration over keys to avoid blocking on large datasets.
