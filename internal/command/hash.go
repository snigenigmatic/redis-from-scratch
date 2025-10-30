package command

import (
	"fmt"

	"redis-from-scratch/internal/store"
)

type HSetHandler struct{}

func (h *HSetHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 3 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'hset' command")}
	}
	key := args[0]
	field := args[1]
	value := args[2]

	n, err := s.HashSet(key, field, value)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: n}
}

type HGetHandler struct{}

func (h *HGetHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'hget' command")}
	}
	key := args[0]
	field := args[1]

	val, ok, err := s.HashGet(key, field)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	if !ok {
		return Response{Type: TypeNull}
	}
	return Response{Type: TypeBulkString, Value: val}
}

type HDelHandler struct{}

func (h *HDelHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'hdel' command")}
	}
	key := args[0]
	fields := args[1:]
	n, err := s.HashDel(key, fields...)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: n}
}

type HGetAllHandler struct{}

func (h *HGetAllHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'hgetall' command")}
	}
	key := args[0]
	m, err := s.HashGetAll(key)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	// Convert map to array of strings [field1, value1, field2, value2]
	arr := make([]string, 0, len(m)*2)
	for k, v := range m {
		arr = append(arr, k, v)
	}
	return Response{Type: TypeArray, Value: arr}
}
