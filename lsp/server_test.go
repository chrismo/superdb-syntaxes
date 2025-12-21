package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestHelper provides utilities for testing the LSP server
type TestHelper struct {
	server *Server
	input  *bytes.Buffer
	output *bytes.Buffer
}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{
		server: NewServer(),
		input:  &bytes.Buffer{},
		output: &bytes.Buffer{},
	}
}

// SendRequest sends a JSON-RPC request to the server
func (h *TestHelper) SendRequest(id interface{}, method string, params interface{}) error {
	msg := RPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}

	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return err
		}
		msg.Params = paramsBytes
	}

	return writeMessage(h.input, msg)
}

// SendNotification sends a JSON-RPC notification to the server
func (h *TestHelper) SendNotification(method string, params interface{}) error {
	msg := RPCMessage{
		JSONRPC: "2.0",
		Method:  method,
	}

	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return err
		}
		msg.Params = paramsBytes
	}

	return writeMessage(h.input, msg)
}

// ReadResponse reads a response from the server output
func (h *TestHelper) ReadResponse() (*RPCMessage, error) {
	var msg RPCMessage
	reader := bufio.NewReader(h.output)
	rawMsg, err := readMessage(reader)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ProcessRequest processes a single request through the server
func (h *TestHelper) ProcessRequest(id interface{}, method string, params interface{}) (*RPCMessage, error) {
	if err := h.SendRequest(id, method, params); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	// Read the message from input and handle it
	reader := bufio.NewReader(h.input)
	rawMsg, err := readMessage(reader)
	if err != nil {
		return nil, fmt.Errorf("read from input: %w", err)
	}

	response, err := h.server.handleMessage(rawMsg)
	if err != nil {
		return nil, fmt.Errorf("handle message: %w", err)
	}

	if response != nil {
		if err := writeMessage(h.output, response); err != nil {
			return nil, fmt.Errorf("write response: %w", err)
		}
		return h.ReadResponse()
	}

	return nil, nil
}

// ProcessNotification processes a notification through the server
func (h *TestHelper) ProcessNotification(method string, params interface{}) (*RPCMessage, error) {
	if err := h.SendNotification(method, params); err != nil {
		return nil, fmt.Errorf("send notification: %w", err)
	}

	reader := bufio.NewReader(h.input)
	rawMsg, err := readMessage(reader)
	if err != nil {
		return nil, fmt.Errorf("read from input: %w", err)
	}

	response, err := h.server.handleMessage(rawMsg)
	if err != nil {
		return nil, fmt.Errorf("handle message: %w", err)
	}

	if response != nil {
		if err := writeMessage(h.output, response); err != nil {
			return nil, fmt.Errorf("write response: %w", err)
		}
		return h.ReadResponse()
	}

	return nil, nil
}

func TestInitializeHandshake(t *testing.T) {
	h := NewTestHelper()

	// Send initialize request
	params := InitializeParams{
		ProcessID: 1234,
		RootURI:   "file:///test",
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Completion: CompletionClientCapabilities{
					CompletionItem: CompletionItemClientCapabilities{
						SnippetSupport: true,
					},
				},
			},
		},
	}

	response, err := h.ProcessRequest(1, "initialize", params)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.ID != float64(1) {
		t.Errorf("Expected ID 1, got %v", response.ID)
	}

	// Check result
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var result InitializeResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("Unmarshal result: %v", err)
	}

	if result.ServerInfo == nil || result.ServerInfo.Name != "superdb-lsp" {
		t.Error("Expected server info with name 'superdb-lsp'")
	}

	if result.Capabilities.TextDocumentSync != 1 {
		t.Errorf("Expected TextDocumentSync 1, got %d", result.Capabilities.TextDocumentSync)
	}

	if result.Capabilities.CompletionProvider == nil {
		t.Error("Expected completion provider")
	}
}

func TestShutdown(t *testing.T) {
	h := NewTestHelper()

	// Initialize first
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Shutdown
	response, err := h.ProcessRequest(2, "shutdown", nil)
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !h.server.shutdown {
		t.Error("Expected server shutdown flag to be true")
	}
}

func TestDidOpenWithValidDocument(t *testing.T) {
	h := NewTestHelper()

	// Initialize first
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open a valid document
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "from test | count()",
		},
	}

	response, err := h.ProcessNotification("textDocument/didOpen", params)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Should receive diagnostics notification
	if response == nil {
		t.Fatal("Expected diagnostics notification, got nil")
	}

	if response.Method != "textDocument/publishDiagnostics" {
		t.Errorf("Expected publishDiagnostics, got %s", response.Method)
	}
}

func TestDidOpenWithInvalidDocument(t *testing.T) {
	h := NewTestHelper()

	// Initialize first
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open an invalid document
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "from test | invalid syntax here {{{{",
		},
	}

	response, err := h.ProcessNotification("textDocument/didOpen", params)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected diagnostics notification, got nil")
	}

	// Parse the diagnostics
	paramsBytes, err := json.Marshal(response.Params)
	if err != nil {
		t.Fatalf("Marshal params: %v", err)
	}

	var diagParams PublishDiagnosticsParams
	if err := json.Unmarshal(paramsBytes, &diagParams); err != nil {
		t.Fatalf("Unmarshal params: %v", err)
	}

	if len(diagParams.Diagnostics) == 0 {
		t.Error("Expected at least one diagnostic for invalid syntax")
	}
}

func TestCompletion(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open document
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "from test | cou",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request completion
	compParams := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Position:     Position{Line: 0, Character: 15},
	}

	response, err := h.ProcessRequest(2, "textDocument/completion", compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected completion response, got nil")
	}

	// Parse completion list
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var completions CompletionList
	if err := json.Unmarshal(resultBytes, &completions); err != nil {
		t.Fatalf("Unmarshal completions: %v", err)
	}

	// Check that "count" is in the completions
	found := false
	for _, item := range completions.Items {
		if item.Label == "count" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'count' in completion items")
	}
}

func TestCompletionKeywords(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open document
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "con",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request completion
	compParams := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Position:     Position{Line: 0, Character: 3},
	}

	response, err := h.ProcessRequest(2, "textDocument/completion", compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected completion response, got nil")
	}

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var completions CompletionList
	if err := json.Unmarshal(resultBytes, &completions); err != nil {
		t.Fatalf("Unmarshal completions: %v", err)
	}

	// Check that "const" is in the completions
	found := false
	for _, item := range completions.Items {
		if item.Label == "const" {
			found = true
			if item.Kind != CompletionItemKindKeyword {
				t.Errorf("Expected keyword kind for 'const', got %d", item.Kind)
			}
			break
		}
	}

	if !found {
		t.Error("Expected 'const' in completion items")
	}
}

func TestCompletionTypes(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open document
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "cast(x, int",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request completion
	compParams := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Position:     Position{Line: 0, Character: 11},
	}

	response, err := h.ProcessRequest(2, "textDocument/completion", compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected completion response, got nil")
	}

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var completions CompletionList
	if err := json.Unmarshal(resultBytes, &completions); err != nil {
		t.Fatalf("Unmarshal completions: %v", err)
	}

	// Check that "int64" is in the completions
	found := false
	for _, item := range completions.Items {
		if item.Label == "int64" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'int64' in completion items")
	}
}

func TestDiagnosticsParseErrors(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		hasError bool
	}{
		{"valid query", "from test | count()", false},
		{"valid with sort", "from test | sort x", false},
		{"invalid syntax", "from {{{{", true},
		{"incomplete pipe", "from test |", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := parseAndGetDiagnostics(tt.code)
			hasError := len(diagnostics) > 0

			if hasError != tt.hasError {
				t.Errorf("Expected hasError=%v, got %v (diagnostics: %v)",
					tt.hasError, hasError, diagnostics)
			}
		})
	}
}

func TestDocumentManagement(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	uri := "file:///test.spq"

	// Open document
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: "spq",
			Version:    1,
			Text:       "from test",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Check document is stored
	if _, ok := h.server.documents[uri]; !ok {
		t.Error("Document not stored after didOpen")
	}

	// Change document
	changeParams := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{URI: uri},
			Version:                2,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: "from test | count()"},
		},
	}
	_, err = h.ProcessNotification("textDocument/didChange", changeParams)
	if err != nil {
		t.Fatalf("didChange failed: %v", err)
	}

	// Check document is updated
	if h.server.documents[uri] != "from test | count()" {
		t.Errorf("Document not updated after didChange: %s", h.server.documents[uri])
	}

	// Close document
	closeParams := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	}
	_, err = h.ProcessNotification("textDocument/didClose", closeParams)
	if err != nil {
		t.Fatalf("didClose failed: %v", err)
	}

	// Check document is removed
	if _, ok := h.server.documents[uri]; ok {
		t.Error("Document not removed after didClose")
	}
}

func TestPositionExtraction(t *testing.T) {
	tests := []struct {
		errStr       string
		expectedLine int
		expectedCol  int
	}{
		{"error at line 5, column 10: syntax error", 4, 9},
		{"line 1:5: unexpected token", 0, 4},
		{"3:7: something went wrong", 2, 6},
		{"no position info", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.errStr, func(t *testing.T) {
			line, col := extractPosition(tt.errStr)
			if line != tt.expectedLine || col != tt.expectedCol {
				t.Errorf("Expected line=%d col=%d, got line=%d col=%d",
					tt.expectedLine, tt.expectedCol, line, col)
			}
		})
	}
}

func TestCompletionPrefixMatching(t *testing.T) {
	tests := []struct {
		text     string
		position Position
		prefix   string
		expected []string
	}{
		{
			text:     "sor",
			position: Position{Line: 0, Character: 3},
			prefix:   "sor",
			expected: []string{"sort"},
		},
		{
			text:     "from test | whe",
			position: Position{Line: 0, Character: 15},
			prefix:   "whe",
			expected: []string{"where"},
		},
		{
			text:     "from test | ",
			position: Position{Line: 0, Character: 12},
			prefix:   "",
			expected: []string{"sort", "where", "count", "yield"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			items := getCompletions(tt.text, tt.position)

			for _, exp := range tt.expected {
				found := false
				for _, item := range items {
					if item.Label == exp {
						found = true
						break
					}
				}
				if !found {
					var labels []string
					for _, item := range items {
						if strings.HasPrefix(item.Label, tt.prefix) {
							labels = append(labels, item.Label)
						}
					}
					t.Errorf("Expected '%s' in completions, got: %v", exp, labels)
				}
			}
		})
	}
}
