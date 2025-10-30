# redis-from-scratch

A small Redis-like server written in Go. It implements a subset of Redis commands
It communicates using the [RESP protocol](https://redis.io/docs/latest/develop/reference/protocol-spec/) and stores data in memory with optional per-key expiry.

This project is meant for learning and experimentation.

Quick start
-----------

1. Build and run the server:

```bash
go run cmd/server/main.go --port 6379
```

2. Test with redis-cli (recommended):

```bash
redis-cli -p 6379 PING
# PONG
redis-cli -p 6379 SET mykey hello
redis-cli -p 6379 GET mykey
# hello
```

If you don't have `redis-cli`, you can use `nc` to send raw RESP messages:

```bash
printf '*1\r\n$4\r\nPING\r\n' | nc -w 1 localhost 6379
# +PONG
```

Key concepts and files
----------------------

- `cmd/server/main.go` - CLI entry point; loads config and starts the server.
- `internal/server` - TCP listener, connection handling, and cleanup loop.
- `internal/protocol` - RESP parser and writer (parsing client requests and writing replies).
- `internal/command` - Command dispatch and command handler implementations (SET, GET, PING, etc.).
- `internal/store` - In-memory key-value store with optional expiry and safe concurrent access.
- `pkg/config` - Default configuration and optional config file loading.

Testing
-------

- Unit tests for the store:

```bash
go test ./internal/store -v
```

- Integration script (starts server, runs commands and checks replies):

```bash
bash scripts/integration_test.sh 6381
```

How to add a command
--------------------

1. Add a new file in `internal/command` implementing the `Handler` interface:

```go
type MyCmdHandler struct{}
func (h *MyCmdHandler) Execute(s *store.Store, args []string) Response { ... }
```

2. Register it in the `handlers` map in `internal/command/command.go` with the uppercase command name.

## TODOs

- [x] Add support for additional data types: hashes (HSET/HGET), lists (LPUSH/LRANGE), sets (SADD/SMEMBERS), and sorted sets (ZADD/ZRANGE).
- [x] Add command handlers and store methods for each data type and enforce type checks (return an error when a command is used on the wrong type).
- [x] Extend the integration script to exercise the new data types and to assert type-error cases.
- [x] Improve RESP parser robustness and add tests for malformed input and large bulk strings.
- [x] Implement persistence (AOF or RDB-style snapshot) and wire `pkg/config` persistence settings into the server.
- [x] Apply per-connection read/write timeouts (use `ReadTimeout`/`WriteTimeout` from config) and add tests for timeout behavior.
- [x] Improve `KEYS` pattern matching or add `SCAN` to avoid blocking on large datasets.
- [x] Add server-level integration tests (Go tests that start the server on an ephemeral port and assert RESP replies).