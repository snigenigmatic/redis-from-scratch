package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"redis-from-scratch/internal/store"
	"redis-from-scratch/pkg/config"
)

type Server struct {
	cfg      *config.Config
	store    *store.Store
	listener net.Listener
	wg       sync.WaitGroup
	quit     chan struct{}
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:   cfg,
		store: store.New(),
		quit:  make(chan struct{}),
	}

	go s.cleanupLoop()

	return s
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.listener = listener
	log.Printf("Redis server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	log.Println("Server stopped")
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
