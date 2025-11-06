package command

import (
	"fmt"

	"redis-from-scratch/internal/store"
)

type SAddHandler struct{}

func (h *SAddHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'sadd' command")}
	}
	key := args[0]
	members := args[1:]
	n, err := s.SetAdd(key, members...)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: n}
}

type SMembersHandler struct{}

func (h *SMembersHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 1 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'smembers' command")}
	}
	key := args[0]
	members, err := s.SetMembers(key)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeArray, Value: members}
}

type SRemHandler struct{}

func (h *SRemHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'srem' command")}
	}
	key := args[0]
	members := args[1:]
	n, err := s.SetRemove(key, members...)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: n}
}

type SISMemberHandler struct{}

func (h *SISMemberHandler) Execute(s *store.Store, args []string) Response {
	if len(args) < 2 {
		return Response{Type: TypeError, Error: fmt.Errorf("ERR : wrong number of arguments for 'sismember' command")}
	}
	key := args[0]
	member := args[1]
	ok, err := s.SetIsMember(key, member)
	if err != nil {
		return Response{Type: TypeError, Error: err}
	}
	return Response{Type: TypeInteger, Value: boolToInt(ok)}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
