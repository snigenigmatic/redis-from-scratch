package protocol

import (
	"strings"
	"testing"
)

func TestParseBasicArray(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 2 || args[0] != "GET" || args[1] != "key" {
		t.Fatalf("expected ['GET', 'key'], got %v", args)
	}
}

func TestParseInlineCommand(t *testing.T) {
	input := "PING\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 1 || args[0] != "PING" {
		t.Fatalf("expected ['PING'], got %v", args)
	}
}

func TestParseMalformedArrayLength(t *testing.T) {
	input := "*abc\r\n"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for malformed array length")
	}
}

func TestParseNegativeArrayLength(t *testing.T) {
	input := "*-5\r\n"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for negative array length")
	}
}

func TestParseArrayTooLarge(t *testing.T) {
	input := "*2000000\r\n"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for array length exceeding limit")
	}
}

func TestParseMissingBulkStringCRLF(t *testing.T) {
	input := "*1\r\n$3\r\nGET"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for missing CRLF")
	}
}

func TestParseInvalidBulkStringLength(t *testing.T) {
	input := "*1\r\n$xyz\r\nGET\r\n"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for invalid bulk string length")
	}
}

func TestParseNegativeBulkStringLength(t *testing.T) {
	input := "*1\r\n$-5\r\n"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for invalid negative bulk string length")
	}
}

func TestParseNullBulkString(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$-1\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 2 || args[0] != "GET" || args[1] != "" {
		t.Fatalf("expected ['GET', ''], got %v", args)
	}
}

func TestParseMultipleCommands(t *testing.T) {
	inputs := []string{
		"*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
		"*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
	}
	expected := [][]string{
		{"SET", "key", "value"},
		{"GET", "key"},
	}

	for i, input := range inputs {
		parser := NewParser(strings.NewReader(input))
		args, err := parser.Parse()
		if err != nil {
			t.Fatalf("command %d: unexpected error: %v", i, err)
		}
		if len(args) != len(expected[i]) {
			t.Fatalf("command %d: expected %d args, got %d", i, len(expected[i]), len(args))
		}
		for j, arg := range args {
			if arg != expected[i][j] {
				t.Fatalf("command %d arg %d: expected %q, got %q", i, j, expected[i][j], arg)
			}
		}
	}
}

func TestParseLargeBulkString(t *testing.T) {
	largeData := strings.Repeat("x", 1000)
	input := "*1\r\n$1000\r\n" + largeData + "\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 1 || args[0] != largeData {
		t.Fatalf("large bulk string not parsed correctly")
	}
}

func TestParseBulkStringExceedsMax(t *testing.T) {
	parser := NewParser(strings.NewReader(""))
	parser.SetMaxBulkLength(100)
	input := "*1\r\n$1000\r\n" + strings.Repeat("x", 1000) + "\r\n"
	parser = NewParser(strings.NewReader(input))
	parser.SetMaxBulkLength(100)
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for bulk string exceeding max length")
	}
}

func TestParseIncompleteBulkString(t *testing.T) {
	input := "*1\r\n$10\r\nhello"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for incomplete bulk string")
	}
}

func TestParseWrongBulkStringType(t *testing.T) {
	input := "*1\r\n+OK\r\n"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.Parse()
	if err == nil {
		t.Fatal("expected error for wrong bulk string marker")
	}
}

func TestParseEmptyCommand(t *testing.T) {
	input := "*0\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected empty args, got %v", args)
	}
}

func TestParseInlineWithMultipleSpaces(t *testing.T) {
	input := "SET   mykey   myvalue\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 3 || args[0] != "SET" || args[1] != "mykey" || args[2] != "myvalue" {
		t.Fatalf("expected ['SET', 'mykey', 'myvalue'], got %v", args)
	}
}

func TestParseBinaryData(t *testing.T) {
	// Test with binary data containing null bytes
	binaryData := "hello\x00world"
	input := "*1\r\n$11\r\n" + binaryData + "\r\n"
	parser := NewParser(strings.NewReader(input))
	args, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) != 1 || args[0] != binaryData {
		t.Fatalf("binary data not parsed correctly")
	}
}
