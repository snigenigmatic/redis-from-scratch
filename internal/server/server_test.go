package server

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"redis-from-scratch/pkg/config"
)

// Helper to start server on ephemeral port
func startTestServer(t *testing.T) (*Server, int) {
	cfg := &config.Config{
		Port:            0, // Use ephemeral port
		MaxConnections:  1000,
		CleanupInterval: 1 * time.Second,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
	}

	srv := New(cfg)

	// Start server and get assigned port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv.listener = listener
	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			srv.wg.Add(1)
			go srv.handleConnection(conn)
		}
	}()

	return srv, port
}

// Helper to send command and get response
func sendCommand(t *testing.T, port int, args []string) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Send RESP array
	fmt.Fprintf(conn, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(arg), arg)
	}

	// Receive and parse the response
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	return string(resp[:n])
}

func TestServerPing(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	resp := sendCommand(t, port, []string{"PING"})
	if !strings.Contains(resp, "PONG") {
		t.Fatalf("expected PONG, got: %s", resp)
	}
}

func TestServerSetGet(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// SET
	resp := sendCommand(t, port, []string{"SET", "mykey", "hello"})
	if !strings.Contains(resp, "OK") {
		t.Fatalf("SET failed: %s", resp)
	}

	// GET
	resp = sendCommand(t, port, []string{"GET", "mykey"})
	if !strings.Contains(resp, "hello") {
		t.Fatalf("GET failed: %s", resp)
	}
}

func TestServerHashOps(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// HSET
	resp := sendCommand(t, port, []string{"HSET", "myhash", "field1", "value1"})
	if !strings.Contains(resp, ":1") {
		t.Fatalf("HSET failed: %s", resp)
	}

	// HGET
	resp = sendCommand(t, port, []string{"HGET", "myhash", "field1"})
	if !strings.Contains(resp, "value1") {
		t.Fatalf("HGET failed: %s", resp)
	}

	// HGETALL
	resp = sendCommand(t, port, []string{"HGETALL", "myhash"})
	if !strings.Contains(resp, "field1") || !strings.Contains(resp, "value1") {
		t.Fatalf("HGETALL failed: %s", resp)
	}
}

func TestServerListOps(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// LPUSH
	resp := sendCommand(t, port, []string{"LPUSH", "mylist", "a", "b", "c"})
	if !strings.Contains(resp, ":3") {
		t.Fatalf("LPUSH failed: %s", resp)
	}

	// LRANGE
	resp = sendCommand(t, port, []string{"LRANGE", "mylist", "0", "-1"})
	if !strings.Contains(resp, "a") || !strings.Contains(resp, "b") || !strings.Contains(resp, "c") {
		t.Fatalf("LRANGE failed: %s", resp)
	}

	// LPOP
	resp = sendCommand(t, port, []string{"LPOP", "mylist"})
	if !strings.Contains(resp, "c") {
		t.Fatalf("LPOP failed: %s", resp)
	}
}

func TestServerSetOps(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// SADD
	resp := sendCommand(t, port, []string{"SADD", "myset", "member1", "member2"})
	if !strings.Contains(resp, ":2") {
		t.Fatalf("SADD failed: %s", resp)
	}

	// SMEMBERS
	resp = sendCommand(t, port, []string{"SMEMBERS", "myset"})
	if !strings.Contains(resp, "member1") || !strings.Contains(resp, "member2") {
		t.Fatalf("SMEMBERS failed: %s", resp)
	}

	// SISMEMBER
	resp = sendCommand(t, port, []string{"SISMEMBER", "myset", "member1"})
	if !strings.Contains(resp, ":1") {
		t.Fatalf("SISMEMBER failed: %s", resp)
	}
}

func TestServerTypeErrors(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// SET a string key
	sendCommand(t, port, []string{"SET", "stringkey", "value"})

	// Try HSET on string key (should error)
	resp := sendCommand(t, port, []string{"HSET", "stringkey", "field", "value"})
	if !strings.Contains(resp, "WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE error, got: %s", resp)
	}

	// Try LPUSH on string key (should error)
	resp = sendCommand(t, port, []string{"LPUSH", "stringkey", "item"})
	if !strings.Contains(resp, "WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE error, got: %s", resp)
	}

	// Try SADD on string key (should error)
	resp = sendCommand(t, port, []string{"SADD", "stringkey", "member"})
	if !strings.Contains(resp, "WRONGTYPE") {
		t.Fatalf("expected WRONGTYPE error, got: %s", resp)
	}
}

func TestServerKeysPattern(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// Set multiple keys
	sendCommand(t, port, []string{"SET", "user:1", "alice"})
	sendCommand(t, port, []string{"SET", "user:2", "bob"})
	sendCommand(t, port, []string{"SET", "post:1", "hello"})

	// KEYS with pattern
	resp := sendCommand(t, port, []string{"KEYS", "user:*"})
	if !strings.Contains(resp, "user:1") || !strings.Contains(resp, "user:2") {
		t.Fatalf("KEYS pattern failed: %s", resp)
	}

	// KEYS *
	resp = sendCommand(t, port, []string{"KEYS", "*"})
	if !strings.Contains(resp, "user:1") && !strings.Contains(resp, "post:1") {
		t.Fatalf("KEYS * failed: %s", resp)
	}
}

func TestServerScan(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// Set multiple keys
	for i := 0; i < 20; i++ {
		sendCommand(t, port, []string{"SET", fmt.Sprintf("key:%d", i), fmt.Sprintf("value%d", i)})
	}

	// SCAN with cursor 0
	resp := sendCommand(t, port, []string{"SCAN", "0", "COUNT", "5"})
	if !strings.Contains(resp, "key:") {
		t.Fatalf("SCAN failed: %s", resp)
	}
}

func TestServerExpiry(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// SET with PX (milliseconds)
	sendCommand(t, port, []string{"SET", "tempkey", "value", "PX", "100"})

	// GET immediately
	resp := sendCommand(t, port, []string{"GET", "tempkey"})
	if !strings.Contains(resp, "value") {
		t.Fatalf("GET before expiry failed: %s", resp)
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// GET after expiry
	resp = sendCommand(t, port, []string{"GET", "tempkey"})
	if strings.Contains(resp, "value") {
		t.Fatalf("key should have expired: %s", resp)
	}
}

func TestServerMultipleConnections(t *testing.T) {
	srv, port := startTestServer(t)
	defer srv.Stop()
	time.Sleep(100 * time.Millisecond)

	// Open multiple connections
	for i := 0; i < 5; i++ {
		go func(id int) {
			conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
			if err != nil {
				t.Errorf("connection %d failed: %v", id, err)
				return
			}
			defer conn.Close()

			// Send command
			fmt.Fprintf(conn, "*3\r\n$3\r\nSET\r\n$5\r\nkey_%d\r\n$6\r\nvalue_%d\r\n", id, id)

			// Read response
			buf := make([]byte, 1024)
			n, _ := conn.Read(buf)
			if !strings.Contains(string(buf[:n]), "OK") {
				t.Errorf("connection %d response wrong: %s", id, buf[:n])
			}
		}(i)
	}

	time.Sleep(500 * time.Millisecond)
}
