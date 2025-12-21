package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// LSP Server for SuperSQL (SPQ) language
// Provides diagnostics and completion support using brimdata/super/compiler

func main() {
	log.SetOutput(os.Stderr)
	log.Println("SuperSQL LSP server starting...")

	server := NewServer()
	if err := server.Run(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Server represents the LSP server
type Server struct {
	documents  map[string]string // URI -> content
	shutdown   bool
	initialized bool
}

// NewServer creates a new LSP server instance
func NewServer() *Server {
	return &Server{
		documents: make(map[string]string),
	}
}

// Run starts the server's main loop
func (s *Server) Run(in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)

	for {
		msg, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("reading message: %w", err)
		}

		response, err := s.handleMessage(msg)
		if err != nil {
			log.Printf("Error handling message: %v", err)
			continue
		}

		if response != nil {
			if err := writeMessage(out, response); err != nil {
				return fmt.Errorf("writing response: %w", err)
			}
		}
	}
}

// readMessage reads a JSON-RPC message from the LSP protocol
func readMessage(reader *bufio.Reader) (json.RawMessage, error) {
	// Read headers
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)

		if line == "" {
			break
		}

		if strings.HasPrefix(line, "Content-Length:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, err = strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("parsing content length: %w", err)
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no content length header")
	}

	// Read content
	content := make([]byte, contentLength)
	_, err := io.ReadFull(reader, content)
	if err != nil {
		return nil, fmt.Errorf("reading content: %w", err)
	}

	return content, nil
}

// writeMessage writes a JSON-RPC message to the output
func writeMessage(out io.Writer, msg interface{}) error {
	content, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(content))
	if _, err := out.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := out.Write(content); err != nil {
		return err
	}
	return nil
}

// handleMessage dispatches incoming JSON-RPC messages
func (s *Server) handleMessage(rawMsg json.RawMessage) (interface{}, error) {
	var msg RPCMessage
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return nil, err
	}

	log.Printf("Received: method=%s, id=%v", msg.Method, msg.ID)

	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		s.initialized = true
		return nil, nil
	case "shutdown":
		return s.handleShutdown(msg)
	case "exit":
		if s.shutdown {
			os.Exit(0)
		}
		os.Exit(1)
	case "textDocument/didOpen":
		return s.handleDidOpen(msg)
	case "textDocument/didChange":
		return s.handleDidChange(msg)
	case "textDocument/didClose":
		return s.handleDidClose(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	default:
		log.Printf("Unhandled method: %s", msg.Method)
	}

	return nil, nil
}
