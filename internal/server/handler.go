package server

import (
	"io"
	"log"
	"net"
	"strings"
	"time"

	"redis-from-scratch/internal/command"
	"redis-from-scratch/internal/protocol"
	"redis-from-scratch/pkg/config"
)

// HandleConnectionWithTimeouts processes client connections with read/write timeouts
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	// Apply read/write timeouts from config
	if err := applyTimeouts(conn, s.cfg); err != nil {
		log.Printf("Warning: failed to apply timeouts: %v", err)
	}

	parser := protocol.NewParser(conn)
	writer := protocol.NewWriter(conn)

	for {
		select {
		case <-s.quit:
			return
		default:
		}

		// Parse incoming command
		args, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("Parse error: %v", err)
			writer.WriteError(err.Error())
			continue
		}

		if len(args) == 0 {
			continue
		}

		cmd := strings.ToUpper(args[0])

		// Execute command
		response := command.Execute(s.store, cmd, args[1:])

		// Persist write commands if persistence enabled
		if s.aof != nil && isPersistentCommand(cmd) {
			if err := s.aof.LogCommand(cmd, args[1:]); err != nil {
				log.Printf("Failed to log command to AOF: %v", err)
				// Don't fail the request, but log the error
			}
		}

		// Write response
		if err := response.WriteTo(writer); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

// applyTimeouts sets read/write deadlines on the connection
func applyTimeouts(conn net.Conn, cfg *config.Config) error {
	if cfg.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout)); err != nil {
			return err
		}
	}

	if cfg.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(cfg.WriteTimeout)); err != nil {
			return err
		}
	}

	return nil
}

// isPersistentCommand determines if a command should be persisted to AOF
func isPersistentCommand(cmd string) bool {
	persistentCommands := map[string]bool{
		"SET":     true,
		"DEL":     true,
		"HSET":    true,
		"HDEL":    true,
		"LPUSH":   true,
		"RPUSH":   true,
		"LPOP":    true,
		"RPOP":    true,
		"SADD":    true,
		"SREM":    true,
		"ZADD":    true,
		"ZREM":    true,
		"FLUSHDB": true,
	}
	return persistentCommands[cmd]
}

// ReadOnlyCommand checks if a command only reads data
func IsReadOnlyCommand(cmd string) bool {
	readOnlyCommands := map[string]bool{
		"GET":       true,
		"HGET":      true,
		"HGETALL":   true,
		"LRANGE":    true,
		"LPOP":      true,
		"RPOP":      true,
		"SMEMBERS":  true,
		"SISMEMBER": true,
		"KEYS":      true,
		"SCAN":      true,
		"HSCAN":     true,
		"EXISTS":    true,
		"PING":      true,
		"ECHO":      true,
	}
	return readOnlyCommands[cmd]
}

// OptimizedHandler with batching and connection pooling consideration
type OptimizedHandler struct {
	commandBuffer []string
	bufferSize    int
}

// Note: For production use, consider:
// 1. Command pipelining (batch multiple commands)
// 2. Connection pooling on client side
// 3. Implementing SELECT for multiple databases
// 4. Rate limiting per connection
// 5. Memory usage limits per key
