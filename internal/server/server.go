package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"redis-from-scratch/internal/command"
	"redis-from-scratch/internal/persistence"
	"redis-from-scratch/internal/store"
	"redis-from-scratch/pkg/config"
)

type Server struct {
	cfg      *config.Config
	store    *store.Store
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
	aof      *persistence.AOF
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:   cfg,
		store: store.New(),
		quit:  make(chan struct{}),
	}

	// Initialize AOF if enabled
	if cfg.EnablePersistence {
		aof, err := persistence.New(cfg.PersistencePath, true)
		if err != nil {
			log.Printf("Warning: failed to initialize AOF: %v", err)
		} else {
			s.aof = aof
			// Replay commands from AOF
			entries, err := aof.ReadCommands()
			if err != nil {
				log.Printf("Warning: failed to read AOF: %v", err)
			} else {
				replayCommands(s.store, entries)
			}
		}
	}

	go s.cleanupLoop()
	return s
}

func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	if s.aof != nil {
		s.aof.Close()
	}
	s.wg.Wait()
	log.Println("Server stopped")
}

func replayCommands(s *store.Store, entries []persistence.AOFEntry) {
	for _, e := range entries {
		// Use command.Execute to replay
		command.Execute(s, e.Command, e.Args)
	}
}

func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(s.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count := s.store.CleanupExpired()
			if count > 0 {
				log.Printf("Cleaned up %d expired keys", count)
			}
		case <-s.quit:
			return
		}
	}
}

// Start begins listening on the configured port and accepts connections.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = ln

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					log.Printf("accept error: %v", err)
					continue
				}
			}
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}()

	return nil
}
