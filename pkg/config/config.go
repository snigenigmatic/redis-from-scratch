package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Port              int           `json:"port"`
	MaxConnections    int           `json:"max_connections"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	MaxRequestSize    int64         `json:"max_request_size"`
	EnablePersistence bool          `json:"enable_persistence"`
	PersistencePath   string        `json:"persistence_path"`
}

func DefaultConfig() *Config {
	return &Config{
		Port:              6379,
		MaxConnections:    1000,
		CleanupInterval:   time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		MaxRequestSize:    512 * 1024 * 1024, // 512MB
		EnablePersistence: false,
		PersistencePath:   "./data",
	}
}

// TODO: The config file includes persistence and timeout fields but the server
// does not currently use persistence or per-connection timeouts. Consider adding
// persistence implementation and applying ReadTimeout/WriteTimeout on connections
// (conn.SetReadDeadline / conn.SetWriteDeadline) in `handleConnection`.

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
