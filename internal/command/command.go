package command

import (
	"fmt"
	"strings"

	"redis-from-scratch/internal/protocol"
	"redis-from-scratch/internal/store"
)

type Handler interface {
	Execute(store *store.Store, args []string) Response
}

type Response struct {
	Type  ResponseType
	Value interface{}
	Error error
}

type ResponseType int

const (
	TypeSimpleString ResponseType = iota
	TypeBulkString
	TypeInteger
	TypeArray
	TypeNull
	TypeError
	TypeNestedArray
)

func (r Response) WriteTo(w *protocol.Writer) error {
	switch r.Type {
	case TypeSimpleString:
		return w.WriteSimpleString(r.Value.(string))
	case TypeBulkString:
		return w.WriteBulkString(r.Value.(string))
	case TypeInteger:
		return w.WriteInteger(r.Value.(int))
	case TypeArray:
		return w.WriteArray(r.Value.([]string))
	case TypeNull:
		return w.WriteNull()
	case TypeError:
		return w.WriteError(r.Error.Error())
	case TypeNestedArray:
		// Value should be a map with "cursor" and "keys" fields
		data := r.Value.(map[string]interface{})
		cursor := data["cursor"].(string)
		keys := data["keys"].([]string)
		return w.WriteNestedArray(cursor, keys)
	default:
		return fmt.Errorf("unknown response type")
	}
}

var handlers = map[string]Handler{
	"PING":      &PingHandler{},
	"ECHO":      &EchoHandler{},
	"SET":       &SetHandler{},
	"GET":       &GetHandler{},
	"HSET":      &HSetHandler{},
	"HGET":      &HGetHandler{},
	"HDEL":      &HDelHandler{},
	"HGETALL":   &HGetAllHandler{},
	"LPUSH":     &LPushHandler{},
	"RPUSH":     &RPushHandler{},
	"LPOP":      &LPopHandler{},
	"RPOP":      &RPopHandler{},
	"LRANGE":    &LRangeHandler{},
	"SADD":      &SAddHandler{},
	"SMEMBERS":  &SMembersHandler{},
	"SREM":      &SRemHandler{},
	"SISMEMBER": &SISMemberHandler{},
	"DEL":       &DelHandler{},
	"EXISTS":    &ExistsHandler{},
	"KEYS":      &KeysHandler{},
	"SCAN":      &ScanHandler{},
	"HSCAN":     &HScanHandler{},
	"ZADD":      &ZAddHandler{},
	"ZRANGE":    &ZRangeHandler{},
}

// TODO: Add handlers for other data types (HSET/HGET for hashes, LPUSH/LRANGE for lists,
// SADD/SMEMBERS for sets, ZADD/ZRANGE for sorted sets). Ensure handlers perform
// type checks and return appropriate errors when the key exists with a different type.

func Execute(s *store.Store, cmd string, args []string) Response {
	handler, ok := handlers[strings.ToUpper(cmd)]
	if !ok {
		return Response{
			Type:  TypeError,
			Error: fmt.Errorf("ERR unknown command '%s'", cmd),
		}
	}
	return handler.Execute(s, args)
}
