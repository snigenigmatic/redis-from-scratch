package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
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
	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %w", err)
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		bulkLine, err := p.readLine()
		if err != nil {
			return nil, err
		}

		if bulkLine[0] != '$' {
			return nil, fmt.Errorf("expected bulk string")
		}

		length, err := strconv.Atoi(bulkLine[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %w", err)
		}

		if length == -1 {
			args[i] = ""
			continue
		}

		buf := make([]byte, length+2)
		if _, err := io.ReadFull(p.reader, buf); err != nil {
			return nil, err
		}

		args[i] = string(buf[:length])
	}

	return args, nil
}

func (p *Parser) parseInline(line string) ([]string, error) {
	return strings.Fields(line), nil
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
