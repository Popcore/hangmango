package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/Popcore/hangmango/pkg/server/handlers"
	"github.com/Popcore/hangmango/pkg/store"
)

// Server is
type Server struct {
	Port    string
	Verbose bool
	Logger  *log.Logger
	System  handlers.System
}

// New returns a fully configured tcp server instance that can be
// used to serve games to players
func New(port string, verbose bool) *Server {
	logger := newLogger(verbose)

	memStore := store.NewMemStore()
	system := handlers.System{
		Store:  memStore,
		Logger: logger,
	}

	return &Server{
		Port:    port,
		Verbose: false,
		Logger:  logger,
		System:  system,
	}
}

func newLogger(verbose bool) *log.Logger {
	if verbose {
		return log.New(os.Stdout, "event: ", log.LstdFlags)
	}

	return log.New(ioutil.Discard, "event: ", log.LstdFlags)
}

// Start listens and responds to incoming client connections.
// Each connection will be managed in its own goroutine.
func (s Server) Start() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer l.Close()

	log.Printf("server listening at port %s", s.Port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		go s.handleConnection(conn)
	}
}

// handleConnection starts a new game session when a new clients connect to the
// server.
func (s Server) handleConnection(conn net.Conn) {
	s.Logger.Println("new client connected")

	err := handlers.NewSession(s.System, conn)
	if err != nil {
		if err == io.EOF {
			s.Logger.Printf("Client disconnected")
		} else {
			s.Logger.Printf("Internal error: %v", err)
		}
	}

	defer conn.Close()
}
