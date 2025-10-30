package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AOF (Append-Only File) persistence implementation
type AOF struct {
	mu       sync.Mutex
	file     *os.File
	writer   *bufio.Writer
	path     string
	enabled  bool
	syncFreq time.Duration
	lastSync time.Time
}

// AOFEntry represents a single command entry in the AOF
type AOFEntry struct {
	Timestamp int64    `json:"ts"`
	Command   string   `json:"cmd"`
	Args      []string `json:"args"`
}

// New creates a new AOF persistence layer
func New(dirPath string, enabled bool) (*AOF, error) {
	if !enabled {
		return &AOF{enabled: false}, nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create persistence directory: %w", err)
	}

	filePath := filepath.Join(dirPath, "commands.aof")

	// Open or create file in append mode
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}

	aof := &AOF{
		file:     f,
		writer:   bufio.NewWriter(f),
		path:     filePath,
		enabled:  true,
		syncFreq: 1 * time.Second,
		lastSync: time.Now(),
	}

	return aof, nil
}

// LogCommand appends a command to the AOF
func (a *AOF) LogCommand(cmd string, args []string) error {
	if !a.enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	entry := AOFEntry{
		Timestamp: time.Now().UnixNano(),
		Command:   cmd,
		Args:      args,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Write JSON + newline
	if _, err := a.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to AOF: %w", err)
	}

	// Periodically sync to disk
	if time.Since(a.lastSync) >= a.syncFreq {
		if err := a.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush AOF: %w", err)
		}
		if err := a.file.Sync(); err != nil {
			return fmt.Errorf("failed to sync AOF: %w", err)
		}
		a.lastSync = time.Now()
	}

	return nil
}

// ReadCommands reads all commands from the AOF file
func (a *AOF) ReadCommands() ([]AOFEntry, error) {
	if !a.enabled {
		return []AOFEntry{}, nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Flush before reading
	if err := a.writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush AOF: %w", err)
	}

	f, err := os.Open(a.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []AOFEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}
	defer f.Close()

	var entries []AOFEntry
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry AOFEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Log malformed line but continue
			fmt.Printf("warning: skipping malformed AOF line %d: %v\n", lineNum, err)
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading AOF file: %w", err)
	}

	return entries, nil
}

// Fsync forces a sync to disk
func (a *AOF) Fsync() error {
	if !a.enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush AOF: %w", err)
	}

	if err := a.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync AOF: %w", err)
	}

	a.lastSync = time.Now()
	return nil
}

// Close closes the AOF file
func (a *AOF) Close() error {
	if !a.enabled || a.file == nil {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush AOF on close: %w", err)
	}

	return a.file.Close()
}

// Truncate clears the AOF file (useful for snapshots)
func (a *AOF) Truncate() error {
	if !a.enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate AOF: %w", err)
	}

	if _, err := a.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek in AOF: %w", err)
	}

	a.writer.Reset(a.file)
	a.lastSync = time.Now()
	return nil
}
