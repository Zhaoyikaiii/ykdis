package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func NewServer(address string) (server *Server, err error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return
	}

	server = &Server{
		listener: listener,
	}

	return
}

func (s *Server) RunAndServe() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	go s.acceptConnectionsLoop()

	<-ctx.Done()

	log.Println("Received shutdown signal, closing server...")

	if err := s.listener.Close(); err != nil {
		log.Println("Error closing listener:", err)
	}
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	<-done
	log.Println("Server shutdown complete")
	return
}

func (s *Server) acceptConnectionsLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("Listener closed, stopping accept loop")
				return
			}
			return
		}
		s.wg.Add(1)
		go s.processConnectionLoop(conn)
	}
}

func (s *Server) processConnectionLoop(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
		log.Println("Connection closed, stopping process loop, address:", conn.RemoteAddr())
	}()

	log.Println("Received new connection from:", conn.RemoteAddr())
	respReader := NewRespReader(conn)
	for {
		args, err := respReader.Args()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("Connection closed by client:", conn.RemoteAddr())
				return
			}
			log.Println("Error parsing command:", err)
			continue
		}
		log.Println("Received command:", args)
		if len(args) == 0 {
			continue
		}
		cmdName := strings.ToUpper(string(args[0]))
		log.Println("Received command:", cmdName)
	}
}
