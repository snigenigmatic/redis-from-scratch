package server

import (
	"io"
	"log"
	"net"
	"strings"

	"redis-from-scratch/internal/command"
	"redis-from-scratch/internal/protocol"
)

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	parser := protocol.NewParser(conn)
	writer := protocol.NewWriter(conn)

	for {
		select {
		case <-s.quit:
			return
		default:
		}

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
		response := command.Execute(s.store, cmd, args[1:])

		if err := response.WriteTo(writer); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}
