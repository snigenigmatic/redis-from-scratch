package command

import (
	"fmt"
	"strconv"

	"redis-from-scratch/internal/store"
)

// ZADD handler: supports one or more score/member pairs.
// Usage: ZADD key score member [score member ...]
type ZAddHandler struct{}

func (h *ZAddHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 3 || ((len(args)-1)%2) != 0 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'zadd' command")}
	}
	key := args[0]

	addedTotal := 0
	for i := 1; i < len(args); i += 2 {
		scoreStr := args[i]
		member := args[i+1]

		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			return Response{Type: TypeError, Error: fmt.Errorf("ERR value is not a valid float")}
		}

		n, err := s.ZAdd(key, score, member)
		if err != nil {
			return Response{Type: TypeError, Error: err}
		}
		addedTotal += n
	}

	return Response{Type: TypeInteger, Value: addedTotal}
}

// ZRANGE handler: ZRANGE key start stop
type ZRangeHandler struct{}

func (h *ZRangeHandler) Execute(s *store.Store, args []string) Response {
	if len(args) != 3 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR wrong number of arguments for 'zrange' command")}
	}
	key := args[0]

	// parse start and stop as integers
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid start index")}
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR invalid stop index")}
	}

	arr, err := s.ZRange(key, start, stop)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeArray, Value: arr}
}
