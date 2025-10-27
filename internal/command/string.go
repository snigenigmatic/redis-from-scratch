package command

import (
	"fmt"
	"strconv"
	"strings"

	"redis-from-scratch/internal/store"
)

type PingHandler struct{}

func (h *PingHandler) Execute(s *store.Store, args []string) Response {
	if len(args) == 0 {
		return Response{Type: TypeSimpleString, Value: "PONG"}
	}
	return Response{Type: TypeBulkString, Value: args[0]}
}

type EchoHandler struct{}

func (h *EchoHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'echo' command")}
	}
	return Response{Type: TypeBulkString, Value: args[0]}
}

type SetHandler struct{}

func (h *SetHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'set' command")}
	}

	key, value := args[0], args[1]
	var expireMs int64

	for i := 2; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return Response{Type: TypeError, Error: fmt.Errorf("ERR syntax error")}
		}

		option := strings.ToUpper(args[i])
		switch option {
		case "PX":
			var err error
			expireMs, err = strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid expire time")}
			}
		case "EX":
			seconds, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid expire time")}
			}
			expireMs = seconds * 1000
		}
	}

	s.Set(key, value, expireMs)
	return Response{Type: TypeSimpleString, Value: "OK"}
}

type GetHandler struct{}

func (h *GetHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'get' command")}
	}

	value, ok := s.Get(args[0])
	if !ok {
		return Response{Type: TypeNull}
	}
	return Response{Type: TypeBulkString, Value: value}
}

// TODO: Add handlers for hash/list/set/zset commands in separate files.
// For example, create `hash.go` with HSET/HGET/HDEL and corresponding store methods.
