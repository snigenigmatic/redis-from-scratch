package command

import (
	"fmt"

	"redis-from-scratch/internal/store"
)

type LPushHandler struct{}

func (h *LPushHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'lpush' command")}
	}
	key := args[0]
	values := args[1:]
	n, err := s.ListLPush(key, values...)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: n}
}

type RPushHandler struct{}

func (h *RPushHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'rpush' command")}
	}
	key := args[0]
	values := args[1:]
	n, err := s.ListRPush(key, values...)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: n}
}

type LPopHandler struct{}

func (h *LPopHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'lpop' command")}
	}
	key := args[0]
	val, ok, err := s.ListLPop(key)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	if !ok {
		return Response{Type: TypeNull}
	}
	return Response{Type: TypeBulkString, Value: val}
}

type RPopHandler struct{}

func (h *RPopHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'rpop' command")}
	}
	key := args[0]
	val, ok, err := s.ListRPop(key)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	if !ok {
		return Response{Type: TypeNull}
	}
	return Response{Type: TypeBulkString, Value: val}
}

type LRangeHandler struct{}

func (h *LRangeHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 3 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'lrange' command")}
	}
	key := args[0]
	// parse start and stop
	var start, stop int
	_, err := fmt.Sscanf(args[1], "%d", &start)
	if err != nil {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : invalid start index")}
	}
	_, err = fmt.Sscanf(args[2], "%d", &stop)
	if err != nil {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : invalid stop index")}
	}
	arr, err := s.ListRange(key, start, stop)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeArray, Value: arr}
}
