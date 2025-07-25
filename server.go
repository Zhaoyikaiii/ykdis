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
	c := newConnectionHandler(conn)
	go func() {
		if err := c.listenCmdLoop(); err != nil {
			log.Println("Error in listenCmdLoop for connection", conn.RemoteAddr(), ":", err)
		}
	}()
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
		cmd := Command{
			Name: strings.ToUpper(args[0]),
			Args: args[1:],
		}
		select {
		case c.cmdCh <- cmd:
		default:
			log.Println("Command channel full, dropping command:", cmd.Name, "from", conn.RemoteAddr())
		}
	}
}

type ConnectionHandler struct {
	conn  net.Conn
	cmdCh chan Command
}

func newConnectionHandler(conn net.Conn) *ConnectionHandler {
	return &ConnectionHandler{
		conn:  conn,
		cmdCh: make(chan Command, PerConnectionCmdBufferSize),
	}
}

type Command struct {
	Name string
	Args []string
}

func (c *ConnectionHandler) listenCmdLoop() (err error) {
	for cmd := range c.cmdCh {
		log.Println(cmd.Name, strings.Join(cmd.Args, " "))
		switch cmd.Name {
		case "PING":
			_, err = c.conn.Write(ping(cmd.Args))
		case "ECHO":
			_, err = c.conn.Write(echo(cmd.Args))
		case "GET":
			_, err = c.conn.Write(get(cmd.Args))
		case "SET":
			_, err = c.conn.Write(set(cmd.Args))
		default:
			_, err = c.conn.Write([]byte("-ERR unknown command '" + cmd.Name + "'\r\n"))
		}
		if err != nil {
			return
		}
	}
	log.Println("Command channel closed, stopping command loop for connection:", c.conn.RemoteAddr())
	return
}
