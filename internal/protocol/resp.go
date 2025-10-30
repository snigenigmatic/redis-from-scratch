package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// TODO: Parser is basic and works for common RESP patterns. Consider hardening it to
// handle edge cases and invalid input more robustly (large bulk lengths, partial reads,
// malformed bytes). Add tests for malformed RESP inputs.

type Parser struct {
	reader    *bufio.Reader
	maxLength int64
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		reader:    bufio.NewReader(r),
		maxLength: 512 * 1024 * 1024, // 512 MB max length
	}
}

func (p *Parser) Parse() ([]string, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}

	if len(line) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	switch line[0] {
	case '*':
		return p.parseArray(line)
	default:
		return p.parseInline(line)
	}
}

func (p *Parser) parseArray(line string) ([]string, error) {
	if len(line) < 2 {
		return nil, fmt.Errorf("malformed array header")
	}

	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %w", err)
	}

	if count < 0 {
		return nil, fmt.Errorf("negative array length: %d", count)
	}

	if count > 1000000 {
		return nil, fmt.Errorf("array length too large: %d", count)
	}

	args := make([]string, 0, count)
	for i := 0; i < count; i++ {
		bulkLine, err := p.readLine()
		if err != nil {
			return nil, fmt.Errorf("error reading bulk string %d: %w", i, err)
		}

		if len(bulkLine) == 0 {
			return nil, fmt.Errorf("empty bulk string header at index %d", i)
		}

		if bulkLine[0] != '$' {
			return nil, fmt.Errorf("expected bulk string at index %d, got %c", i, bulkLine[0])
		}

		length, err := strconv.ParseInt(bulkLine[1:], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length at index %d: %w", i, err)
		}

		if length < -1 {
			return nil, fmt.Errorf("invalid bulk string length at index %d: %d", i, length)
		}

		if length == -1 {
			// Null bulk string
			args = append(args, "")
			continue
		}

		if length > p.maxLength {
			return nil, fmt.Errorf("bulk string exceeds max length at index %d: %d > %d", i, length, p.maxLength)
		}

		buf := make([]byte, length+2)
		n, err := io.ReadFull(p.reader, buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read bulk string data at index %d: %w (read %d/%d bytes)", i, err, n, length+2)
		}

		if buf[length] != '\r' || buf[length+1] != '\n' {
			return nil, fmt.Errorf("bulk string at index %d missing CRLF terminator", i)
		}

		args = append(args, string(buf[:length]))
	}

	return args, nil
}

func (p *Parser) parseInline(line string) ([]string, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty inline command")
	}
	return parts, nil
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && len(line) > 0 {
			return "", fmt.Errorf("incomplete line: %w", err)
		}
		return "", err
	}

	line = strings.TrimRight(line, "\r\n")

	if len(line) == 0 {
		return "", fmt.Errorf("empty line")
	}

	return line, nil
}

// SetMaxBulkLength sets the maximum allowed bulk string length parsed by the parser.
// This is useful for tests to restrict sizes and for callers to limit memory use.
func (p *Parser) SetMaxBulkLength(n int64) {
	p.maxLength = n
}
