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

func TestCompletionSQLKeywords(t *testing.T) {
	// Test that SQL keywords are available in completions
	sqlKeywords := []string{
		"select", "group", "having", "order", "limit", "offset",
		"join", "left", "right", "inner", "outer", "on",
		"case", "when", "then", "else", "end",
		"and", "or", "not", "in", "like", "between",
	}

	items := getCompletions("", Position{Line: 0, Character: 0})

	for _, kw := range sqlKeywords {
		found := false
		for _, item := range items {
			if item.Label == kw {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SQL keyword '%s' not found in completions", kw)
		}
	}
}

func TestCompletionOperators(t *testing.T) {
	// Test that all operators are available
	ops := []string{
		"assert", "cut", "drop", "fork", "fuse",
		"head", "join", "merge", "over", "pass", "put",
		"rename", "sort", "summarize", "tail", "uniq", "where", "yield",
		"debug", "explode", "output", "skip", "unnest", "values",
	}

	items := getCompletions("", Position{Line: 0, Character: 0})

	for _, op := range ops {
		found := false
		for _, item := range items {
			if item.Label == op {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Operator '%s' not found in completions", op)
		}
	}
}

func TestCompletionFunctions(t *testing.T) {
	// Test that all functions are available
	funcs := []string{
		"abs", "cast", "ceil", "floor", "len", "lower", "upper",
		"split", "trim", "typeof", "coalesce", "has", "grep",
		// New functions from PEG grammar
		"date_part", "length", "nullif", "parse_sup", "position",
	}

	items := getCompletions("test(", Position{Line: 0, Character: 5})

	for _, fn := range funcs {
		found := false
		for _, item := range items {
			if item.Label == fn {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Function '%s' not found in completions", fn)
		}
	}
}

func TestCompletionAggregates(t *testing.T) {
	// Test that all aggregates are available
	aggs := []string{
		"count", "sum", "avg", "min", "max",
		"collect", "collect_map", "dcount", "union", "any", "fuse",
	}

	items := getCompletions("summarize(", Position{Line: 0, Character: 10})

	for _, agg := range aggs {
		found := false
		for _, item := range items {
			if item.Label == agg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Aggregate '%s' not found in completions", agg)
		}
	}
}

func TestCompletionAllTypes(t *testing.T) {
	// Test that all types are available including new ones
	allTypes := []string{
		// Core types
		"int64", "uint64", "float64", "string", "bool", "bytes",
		"time", "duration", "ip", "net", "null", "type",
		// New SQL type aliases
		"date", "timestamp", "bigint", "smallint", "boolean", "text", "bytea",
	}

	items := getCompletions("cast(x, ", Position{Line: 0, Character: 8})

	for _, typ := range allTypes {
		found := false
		for _, item := range items {
			if item.Label == typ {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Type '%s' not found in completions", typ)
		}
	}
}

func TestCompletionContext(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		col      int
		expected completionContext
	}{
		{"general context", "from test", 9, contextGeneral},
		{"type context after cast", "cast(x, ", 8, contextType},
		{"type context after ::", "x::", 3, contextType},
		{"function context in parens", "foo(bar", 7, contextFunction},
		{"general after closed parens", "foo()", 5, contextGeneral},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := getCompletionContext(tt.line, tt.col)
			if ctx != tt.expected {
				t.Errorf("Expected context %d, got %d", tt.expected, ctx)
			}
		})
	}
}

func TestDiagnosticsValidQueries(t *testing.T) {
	// Test various valid query patterns
	validQueries := []string{
		"from test",
		"from test | count()",
		"from test | where x > 5",
		"from test | sort x",
		"from test | head 10",
		"from test | put y := x + 1",
		"from test | summarize count() by x",
		"from test | values {a: 1}",
	}

	for _, query := range validQueries {
		t.Run(query, func(t *testing.T) {
			diagnostics := parseAndGetDiagnostics(query)
			if len(diagnostics) > 0 {
				t.Errorf("Expected no diagnostics for valid query, got: %v", diagnostics[0].Message)
			}
		})
	}
}

func TestDiagnosticsInvalidQueries(t *testing.T) {
	// Test various invalid query patterns
	invalidQueries := []string{
		"from {{{{",
		"from test |",
		"| count()",
		"from test | sort >>>",
	}

	for _, query := range invalidQueries {
		t.Run(query, func(t *testing.T) {
			diagnostics := parseAndGetDiagnostics(query)
			if len(diagnostics) == 0 {
				t.Errorf("Expected diagnostics for invalid query: %s", query)
			}
		})
	}
}

// Tests for migration diagnostics
func TestMigrationDiagnostics(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantCode string
		wantMsg  string
	}{
		{
			name:     "deprecated yield",
			query:    "from test | yield {a: 1}",
			wantCode: "deprecated-yield",
			wantMsg:  "'yield' is deprecated, use 'values'",
		},
		{
			name:     "deprecated func",
			query:    "func add(a, b): a + b",
			wantCode: "deprecated-func",
			wantMsg:  "'func' is deprecated, use 'fn'",
		},
		{
			name:     "deprecated arrow",
			query:    "from test => output",
			wantCode: "deprecated-arrow",
			wantMsg:  "'=>' is deprecated, use 'into'",
		},
		{
			name:     "deprecated parse_zson",
			query:    `parse_zson("{a:1}")`,
			wantCode: "deprecated-parse-zson",
			wantMsg:  "'parse_zson' is deprecated, use 'parse_sup'",
		},
		{
			name:     "deprecated cast time",
			query:    `time('2025-01-01T00:00:00Z')`,
			wantCode: "deprecated-cast-time",
			wantMsg:  "Function-style cast deprecated, use '::time'",
		},
		{
			name:     "implicit this grep",
			query:    `grep(/error/)`,
			wantCode: "implicit-this-grep",
			wantMsg:  "grep() requires explicit 'this' argument",
		},
		{
			name:     "implicit this is",
			query:    `is(<string>)`,
			wantCode: "implicit-this-is",
			wantMsg:  "is() requires explicit 'this' argument",
		},
		{
			name:     "implicit this nest_dotted",
			query:    `nest_dotted()`,
			wantCode: "implicit-this-nest-dotted",
			wantMsg:  "nest_dotted() requires explicit 'this' argument",
		},
		{
			name:     "removed crop",
			query:    `crop(this)`,
			wantCode: "removed-crop",
			wantMsg:  "'crop()' was removed, use explicit casting",
		},
		{
			name:     "deprecated comment slash",
			query:    `from test // this is a comment`,
			wantCode: "deprecated-comment-slash",
			wantMsg:  "'//' comments are deprecated, use '--'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := getMigrationDiagnostics(tt.query)
			if len(diags) == 0 {
				t.Errorf("Expected migration diagnostic for: %s", tt.query)
				return
			}
			found := false
			for _, d := range diags {
				if d.Diagnostic.Code == tt.wantCode {
					found = true
					if d.Diagnostic.Message != tt.wantMsg {
						t.Errorf("Got message %q, want %q", d.Diagnostic.Message, tt.wantMsg)
					}
					break
				}
			}
			if !found {
				t.Errorf("Expected diagnostic code %q, got %v", tt.wantCode, diags[0].Diagnostic.Code)
			}
		})
	}
}

func TestMigrationFixes(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantFix string
	}{
		{
			name:    "fix yield to values",
			query:   "yield x",
			wantFix: "values",
		},
		{
			name:    "fix func to fn",
			query:   "func add(a): a",
			wantFix: "fn",
		},
		{
			name:    "fix arrow to into",
			query:   "from test => out",
			wantFix: "into",
		},
		{
			name:    "fix parse_zson to parse_sup",
			query:   `parse_zson("{a:1}")`,
			wantFix: "parse_sup(",
		},
		{
			name:    "fix time cast",
			query:   `time('2025-01-01')`,
			wantFix: "'2025-01-01'::time",
		},
		{
			name:    "fix grep with this",
			query:   `grep(/error/)`,
			wantFix: "grep('error', this)",
		},
		{
			name:    "fix is with this",
			query:   `is(<string>)`,
			wantFix: "is(this, <string>)",
		},
		{
			name:    "fix nest_dotted with this",
			query:   `nest_dotted()`,
			wantFix: "nest_dotted(this)",
		},
		{
			name:    "fix comment slash to dash",
			query:   `from test // comment`,
			wantFix: " --", // preserves space before comment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := getMigrationDiagnostics(tt.query)
			if len(diags) == 0 {
				t.Errorf("Expected migration diagnostic for: %s", tt.query)
				return
			}
			if diags[0].Fix == nil {
				t.Errorf("Expected fix for: %s", tt.query)
				return
			}
			if diags[0].Fix.NewText != tt.wantFix {
				t.Errorf("Got fix %q, want %q", diags[0].Fix.NewText, tt.wantFix)
			}
		})
	}
}

func TestMigrationMultiLine(t *testing.T) {
	// Test multi-line document with multiple deprecated items
	text := `from test
| yield x
| func add(a): a + 1
| where x > 0 // filter positives`

	diags := getMigrationDiagnostics(text)

	// Should find: yield (line 1), func (line 2), // comment (line 3)
	expectedCodes := map[string]bool{
		"deprecated-yield":         false,
		"deprecated-func":          false,
		"deprecated-comment-slash": false,
	}

	for _, d := range diags {
		if _, ok := expectedCodes[d.Diagnostic.Code]; ok {
			expectedCodes[d.Diagnostic.Code] = true
		}
	}

	for code, found := range expectedCodes {
		if !found {
			t.Errorf("Expected to find diagnostic %q in multi-line text", code)
		}
	}

	// Verify line numbers are correct
	for _, d := range diags {
		switch d.Diagnostic.Code {
		case "deprecated-yield":
			if d.Diagnostic.Range.Start.Line != 1 {
				t.Errorf("yield should be on line 1, got %d", d.Diagnostic.Range.Start.Line)
			}
		case "deprecated-func":
			if d.Diagnostic.Range.Start.Line != 2 {
				t.Errorf("func should be on line 2, got %d", d.Diagnostic.Range.Start.Line)
			}
		case "deprecated-comment-slash":
			if d.Diagnostic.Range.Start.Line != 3 {
				t.Errorf("// should be on line 3, got %d", d.Diagnostic.Range.Start.Line)
			}
		}
	}
}

func TestMigrationMultipleOnSameLine(t *testing.T) {
	// Multiple deprecated items on the same line
	text := `yield x => output // done`

	diags := getMigrationDiagnostics(text)

	// Should find: yield, =>, //
	if len(diags) < 3 {
		t.Errorf("Expected at least 3 diagnostics, got %d", len(diags))
		for _, d := range diags {
			t.Logf("  Found: %s at col %d", d.Diagnostic.Code, d.Diagnostic.Range.Start.Character)
		}
	}

	// All should be on line 0
	for _, d := range diags {
		if d.Diagnostic.Range.Start.Line != 0 {
			t.Errorf("All diagnostics should be on line 0, got line %d for %s",
				d.Diagnostic.Range.Start.Line, d.Diagnostic.Code)
		}
	}
}

func TestMigrationNoFalsePositives(t *testing.T) {
	// These should NOT trigger migration diagnostics
	queries := []string{
		"from test | values {a: 1}",            // modern syntax
		"fn add(a, b): a + b",                  // modern syntax
		"from test into output",                // modern syntax
		`parse_sup("{a:1}")`,                   // modern syntax
		`'2025-01-01'::time`,                   // modern syntax
		"grep('pattern', this)",                // explicit this
		"is(this, <string>)",                   // explicit this
		"nest_dotted(this)",                    // explicit this
		"https://example.com",                  // URL should not match //
		"from test -- this is a comment",       // modern comment syntax
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			diags := getMigrationDiagnostics(query)
			if len(diags) > 0 {
				t.Errorf("Unexpected migration diagnostic for: %s, got code=%s msg=%s",
					query, diags[0].Diagnostic.Code, diags[0].Diagnostic.Message)
			}
		})
	}
}

func TestKeywordCount(t *testing.T) {
	// Verify we have a reasonable number of keywords
	if len(Builtins.Keywords()) < 40 {
		t.Errorf("Expected at least 40 keywords, got %d", len(Builtins.Keywords()))
	}
}

func TestOperatorCount(t *testing.T) {
	// Verify we have a reasonable number of operators
	if len(Builtins.Operators()) < 25 {
		t.Errorf("Expected at least 25 operators, got %d", len(Builtins.Operators()))
	}
}

func TestFunctionCount(t *testing.T) {
	// Verify we have a reasonable number of functions
	if len(Builtins.Functions()) < 50 {
		t.Errorf("Expected at least 50 functions, got %d", len(Builtins.Functions()))
	}
}

func TestTypeCount(t *testing.T) {
	// Verify we have a reasonable number of types
	if len(Builtins.Types()) < 35 {
		t.Errorf("Expected at least 35 types, got %d", len(Builtins.Types()))
	}
}

// Tests for new LSP features

func TestHoverKeyword(t *testing.T) {
	text := "from test | where x > 5"
	pos := Position{Line: 0, Character: 13} // over "where"

	hover := getHover(text, pos)
	if hover == nil {
		t.Fatal("Expected hover result, got nil")
	}

	if hover.Contents.Kind != MarkupKindMarkdown {
		t.Errorf("Expected markdown content, got %s", hover.Contents.Kind)
	}

	if !strings.Contains(hover.Contents.Value, "where") {
		t.Errorf("Expected hover to contain 'where', got: %s", hover.Contents.Value)
	}
}

func TestHoverFunction(t *testing.T) {
	text := "from test | put y := ceil(x)"
	pos := Position{Line: 0, Character: 22} // over "ceil"

	hover := getHover(text, pos)
	if hover == nil {
		t.Fatal("Expected hover result, got nil")
	}

	if !strings.Contains(hover.Contents.Value, "ceil") {
		t.Errorf("Expected hover to contain 'ceil', got: %s", hover.Contents.Value)
	}
}

func TestHoverAggregate(t *testing.T) {
	text := "from test | summarize count() by x"
	pos := Position{Line: 0, Character: 23} // over "count"

	hover := getHover(text, pos)
	if hover == nil {
		t.Fatal("Expected hover result, got nil")
	}

	if !strings.Contains(hover.Contents.Value, "count") {
		t.Errorf("Expected hover to contain 'count', got: %s", hover.Contents.Value)
	}
}

func TestHoverType(t *testing.T) {
	text := "cast(x, int64)"
	pos := Position{Line: 0, Character: 9} // over "int64"

	hover := getHover(text, pos)
	if hover == nil {
		t.Fatal("Expected hover result, got nil")
	}

	if !strings.Contains(hover.Contents.Value, "int64") {
		t.Errorf("Expected hover to contain 'int64', got: %s", hover.Contents.Value)
	}
}

func TestHoverNoResult(t *testing.T) {
	text := "from test"
	pos := Position{Line: 0, Character: 5} // over "test" (not a keyword)

	hover := getHover(text, pos)
	if hover != nil {
		t.Errorf("Expected no hover for identifier, got: %v", hover)
	}
}

func TestSignatureHelpFunction(t *testing.T) {
	text := "from test | put y := ceil("
	pos := Position{Line: 0, Character: 26} // after opening paren

	sigHelp := getSignatureHelp(text, pos)
	if sigHelp == nil {
		t.Fatal("Expected signature help, got nil")
	}

	if len(sigHelp.Signatures) != 1 {
		t.Fatalf("Expected 1 signature, got %d", len(sigHelp.Signatures))
	}

	sig := sigHelp.Signatures[0]
	if !strings.Contains(sig.Label, "ceil") {
		t.Errorf("Expected signature for 'ceil', got: %s", sig.Label)
	}
}

func TestSignatureHelpAggregate(t *testing.T) {
	text := "from test | summarize sum("
	pos := Position{Line: 0, Character: 26}

	sigHelp := getSignatureHelp(text, pos)
	if sigHelp == nil {
		t.Fatal("Expected signature help, got nil")
	}

	if len(sigHelp.Signatures) != 1 {
		t.Fatalf("Expected 1 signature, got %d", len(sigHelp.Signatures))
	}

	sig := sigHelp.Signatures[0]
	if !strings.Contains(sig.Label, "sum") {
		t.Errorf("Expected signature for 'sum', got: %s", sig.Label)
	}
}

func TestSignatureHelpMultipleParams(t *testing.T) {
	text := "replace(s, old, "
	pos := Position{Line: 0, Character: 16} // after second comma

	sigHelp := getSignatureHelp(text, pos)
	if sigHelp == nil {
		t.Fatal("Expected signature help, got nil")
	}

	if sigHelp.ActiveParameter != 2 {
		t.Errorf("Expected active parameter 2, got %d", sigHelp.ActiveParameter)
	}
}

func TestSignatureHelpNoContext(t *testing.T) {
	text := "from test | sort x"
	pos := Position{Line: 0, Character: 18}

	sigHelp := getSignatureHelp(text, pos)
	if sigHelp != nil {
		t.Errorf("Expected no signature help outside function call, got: %v", sigHelp)
	}
}

func TestFormatBasic(t *testing.T) {
	input := "from   test  |   count()"
	expected := "from test\n| count()"

	options := FormattingOptions{
		TabSize:      2,
		InsertSpaces: true,
	}

	result := formatDocument(input, options)
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestFormatPreservesComments(t *testing.T) {
	input := "-- comment\nfrom test"

	options := FormattingOptions{
		TabSize:      2,
		InsertSpaces: true,
	}

	result := formatDocument(input, options)
	if !strings.Contains(result, "-- comment") {
		t.Errorf("Expected comment to be preserved, got: %s", result)
	}
}

func TestFormatPreservesStrings(t *testing.T) {
	input := `from test | put x := "hello   world"`

	options := FormattingOptions{
		TabSize:      2,
		InsertSpaces: true,
	}

	result := formatDocument(input, options)
	if !strings.Contains(result, `"hello   world"`) {
		t.Errorf("Expected string content to be preserved, got: %s", result)
	}
}

func TestFormatPipeOnNewLine(t *testing.T) {
	input := "from test|count()|sort x"

	options := FormattingOptions{
		TabSize:      2,
		InsertSpaces: true,
	}

	result := formatDocument(input, options)
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines (one per pipe), got %d: %s", len(lines), result)
	}
}

func TestFormatWithFinalNewline(t *testing.T) {
	input := "from test"

	options := FormattingOptions{
		TabSize:           2,
		InsertSpaces:      true,
		InsertFinalNewline: true,
	}

	result := formatDocument(input, options)
	if !strings.HasSuffix(result, "\n") {
		t.Errorf("Expected final newline, got: %q", result)
	}
}

func TestFormatTrimTrailingWhitespace(t *testing.T) {
	input := "from test   \n| count()   "

	options := FormattingOptions{
		TabSize:                2,
		InsertSpaces:           true,
		TrimTrailingWhitespace: true,
	}

	result := formatDocument(input, options)
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, " ") {
			t.Errorf("Line has trailing whitespace: %q", line)
		}
	}
}

func TestHoverHandler(t *testing.T) {
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
			Text:       "from test | sort x",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request hover over "sort"
	hoverParams := HoverParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Position:     Position{Line: 0, Character: 13},
	}

	response, err := h.ProcessRequest(2, "textDocument/hover", hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected hover response, got nil")
	}

	// Parse hover result
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var hover Hover
	if err := json.Unmarshal(resultBytes, &hover); err != nil {
		t.Fatalf("Unmarshal hover: %v", err)
	}

	if !strings.Contains(hover.Contents.Value, "sort") {
		t.Errorf("Expected hover to contain 'sort', got: %s", hover.Contents.Value)
	}
}

func TestSignatureHelpHandler(t *testing.T) {
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
			Text:       "from test | put y := ceil(",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request signature help
	sigParams := SignatureHelpParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Position:     Position{Line: 0, Character: 26},
	}

	response, err := h.ProcessRequest(2, "textDocument/signatureHelp", sigParams)
	if err != nil {
		t.Fatalf("SignatureHelp failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected signature help response, got nil")
	}

	// Parse signature help result
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var sigHelp SignatureHelp
	if err := json.Unmarshal(resultBytes, &sigHelp); err != nil {
		t.Fatalf("Unmarshal signature help: %v", err)
	}

	if len(sigHelp.Signatures) == 0 {
		t.Error("Expected at least one signature")
	}
}

func TestFormattingHandler(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open document with messy formatting
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "from   test  |  count()",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request formatting
	formatParams := DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Options: FormattingOptions{
			TabSize:      2,
			InsertSpaces: true,
		},
	}

	response, err := h.ProcessRequest(2, "textDocument/formatting", formatParams)
	if err != nil {
		t.Fatalf("Formatting failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected formatting response, got nil")
	}

	// Parse text edits
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var edits []TextEdit
	if err := json.Unmarshal(resultBytes, &edits); err != nil {
		t.Fatalf("Unmarshal edits: %v", err)
	}

	if len(edits) == 0 {
		t.Error("Expected at least one edit for messy input")
	}
}

func TestCodeActionHandler(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open document with deprecated syntax
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "from test | yield {a: 1}",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request code actions with the diagnostic context
	codeActionParams := CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Range: Range{
			Start: Position{Line: 0, Character: 12},
			End:   Position{Line: 0, Character: 17},
		},
		Context: CodeActionContext{
			Diagnostics: []Diagnostic{
				{
					Range: Range{
						Start: Position{Line: 0, Character: 12},
						End:   Position{Line: 0, Character: 17},
					},
					Severity: DiagnosticSeverityWarning,
					Code:     "deprecated-yield",
					Source:   "superdb-lsp",
					Message:  "'yield' is deprecated, use 'values'",
				},
			},
		},
	}

	response, err := h.ProcessRequest(2, "textDocument/codeAction", codeActionParams)
	if err != nil {
		t.Fatalf("CodeAction failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected code action response, got nil")
	}

	// Parse code actions
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var actions []CodeAction
	if err := json.Unmarshal(resultBytes, &actions); err != nil {
		t.Fatalf("Unmarshal actions: %v", err)
	}

	if len(actions) == 0 {
		t.Fatal("Expected at least one code action")
	}

	// Verify the quick fix action
	found := false
	for _, action := range actions {
		if action.Kind == CodeActionKindQuickFix && strings.Contains(action.Title, "values") {
			found = true
			if action.Edit == nil {
				t.Error("Expected edit in code action")
			} else if edits, ok := action.Edit.Changes["file:///test.spq"]; ok {
				if len(edits) == 0 {
					t.Error("Expected edits for file")
				} else if edits[0].NewText != "values" {
					t.Errorf("Expected NewText 'values', got %q", edits[0].NewText)
				}
			} else {
				t.Error("Expected changes for file:///test.spq")
			}
			break
		}
	}
	if !found {
		t.Error("Expected quick fix action with 'values' replacement")
	}
}

func TestCodeActionFixAll(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open document with multiple deprecated items
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.spq",
			LanguageID: "spq",
			Version:    1,
			Text:       "from test | yield x => output",
		},
	}
	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request code actions - should include "Fix all"
	codeActionParams := CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.spq"},
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 0, Character: 29},
		},
		Context: CodeActionContext{
			Diagnostics: []Diagnostic{
				{
					Range:    Range{Start: Position{Line: 0, Character: 12}, End: Position{Line: 0, Character: 17}},
					Code:     "deprecated-yield",
					Message:  "'yield' is deprecated, use 'values'",
				},
				{
					Range:    Range{Start: Position{Line: 0, Character: 20}, End: Position{Line: 0, Character: 22}},
					Code:     "deprecated-arrow",
					Message:  "'=>' is deprecated, use 'into'",
				},
			},
		},
	}

	response, err := h.ProcessRequest(2, "textDocument/codeAction", codeActionParams)
	if err != nil {
		t.Fatalf("CodeAction failed: %v", err)
	}

	resultBytes, _ := json.Marshal(response.Result)
	var actions []CodeAction
	json.Unmarshal(resultBytes, &actions)

	// Should have individual fixes plus "Fix all"
	hasFixAll := false
	for _, action := range actions {
		if action.Kind == CodeActionKindSourceFixAll {
			hasFixAll = true
			if action.Edit == nil {
				t.Error("Fix all should have edits")
			} else if edits, ok := action.Edit.Changes["file:///test.spq"]; ok {
				if len(edits) < 2 {
					t.Errorf("Fix all should have at least 2 edits, got %d", len(edits))
				}
			}
			break
		}
	}
	if !hasFixAll {
		t.Error("Expected 'Fix all' action when multiple issues present")
	}
}

func TestSupFileSkipsDiagnostics(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open a .sup file with content that would normally produce parse errors
	// .sup files contain data sequences, not queries, so they should not be parsed
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.sup",
			LanguageID: "sup",
			Version:    1,
			Text:       "{name: \"test\", value: 42}\n{name: \"other\", value: 99}",
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

	// .sup files should produce zero diagnostics since parsing is skipped
	if len(diagParams.Diagnostics) != 0 {
		t.Errorf("Expected 0 diagnostics for .sup file, got %d: %v",
			len(diagParams.Diagnostics), diagParams.Diagnostics)
	}
}

func TestSupFileCaseInsensitive(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test with uppercase .SUP extension - using valid data
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.SUP",
			LanguageID: "sup",
			Version:    1,
			Text:       "{name: \"uppercase\", value: 123}",
		},
	}

	response, err := h.ProcessNotification("textDocument/didOpen", params)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	paramsBytes, err := json.Marshal(response.Params)
	if err != nil {
		t.Fatalf("Marshal params: %v", err)
	}

	var diagParams PublishDiagnosticsParams
	if err := json.Unmarshal(paramsBytes, &diagParams); err != nil {
		t.Fatalf("Unmarshal params: %v", err)
	}

	// Valid data should produce no diagnostics (case insensitive check)
	if len(diagParams.Diagnostics) != 0 {
		t.Errorf("Expected 0 diagnostics for valid .SUP file (case insensitive), got %d: %v",
			len(diagParams.Diagnostics), diagParams.Diagnostics)
	}
}

func TestSupFileInvalidDataDiagnostics(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open a .sup file with invalid data syntax
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.sup",
			LanguageID: "sup",
			Version:    1,
			Text:       "{invalid syntax without proper values}",
		},
	}

	response, err := h.ProcessNotification("textDocument/didOpen", params)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	paramsBytes, err := json.Marshal(response.Params)
	if err != nil {
		t.Fatalf("Marshal params: %v", err)
	}

	var diagParams PublishDiagnosticsParams
	if err := json.Unmarshal(paramsBytes, &diagParams); err != nil {
		t.Fatalf("Unmarshal params: %v", err)
	}

	// Invalid data should produce diagnostics
	if len(diagParams.Diagnostics) == 0 {
		t.Error("Expected diagnostics for invalid .sup data, got none")
	}
}

func TestSupFileFormatting(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open a .sup file with unformatted data
	openParams := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.sup",
			LanguageID: "sup",
			Version:    1,
			Text:       "{name:\"test\",value:42}",
		},
	}

	_, err = h.ProcessNotification("textDocument/didOpen", openParams)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	// Request formatting
	formatParams := DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.sup"},
		Options: FormattingOptions{
			TabSize:      4,
			InsertSpaces: true,
		},
	}

	response, err := h.ProcessRequest(2, "textDocument/formatting", formatParams)
	if err != nil {
		t.Fatalf("formatting failed: %v", err)
	}

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var edits []TextEdit
	if err := json.Unmarshal(resultBytes, &edits); err != nil {
		t.Fatalf("Unmarshal edits: %v", err)
	}

	// We should get formatting edits
	if len(edits) == 0 {
		t.Error("Expected formatting edits for .sup file, got none")
	}

	// The formatted output should have proper spacing
	if len(edits) > 0 {
		formatted := edits[0].NewText
		// Check that fields are properly spaced
		if !strings.Contains(formatted, "name:") || !strings.Contains(formatted, "value:") {
			t.Errorf("Formatted output should have proper structure: %s", formatted)
		}
	}
}

func TestSupFileMultipleValues(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open a .sup file with multiple data values
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.sup",
			LanguageID: "sup",
			Version:    1,
			Text: `{name: "first", value: 1}
{name: "second", value: 2}
{name: "third", value: 3}`,
		},
	}

	response, err := h.ProcessNotification("textDocument/didOpen", params)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	paramsBytes, err := json.Marshal(response.Params)
	if err != nil {
		t.Fatalf("Marshal params: %v", err)
	}

	var diagParams PublishDiagnosticsParams
	if err := json.Unmarshal(paramsBytes, &diagParams); err != nil {
		t.Fatalf("Unmarshal params: %v", err)
	}

	// Valid multiple values should produce no diagnostics
	if len(diagParams.Diagnostics) != 0 {
		t.Errorf("Expected 0 diagnostics for valid multi-value .sup file, got %d: %v",
			len(diagParams.Diagnostics), diagParams.Diagnostics)
	}
}

func TestSupFileComplexData(t *testing.T) {
	h := NewTestHelper()

	// Initialize
	_, err := h.ProcessRequest(1, "initialize", InitializeParams{ProcessID: 1})
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Open a .sup file with complex data types
	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.sup",
			LanguageID: "sup",
			Version:    1,
			Text:       `{array: [1, 2, 3], nested: {inner: "value"}, time: 2024-01-15T10:30:00Z}`,
		},
	}

	response, err := h.ProcessNotification("textDocument/didOpen", params)
	if err != nil {
		t.Fatalf("didOpen failed: %v", err)
	}

	paramsBytes, err := json.Marshal(response.Params)
	if err != nil {
		t.Fatalf("Marshal params: %v", err)
	}

	var diagParams PublishDiagnosticsParams
	if err := json.Unmarshal(paramsBytes, &diagParams); err != nil {
		t.Fatalf("Unmarshal params: %v", err)
	}

	// Valid complex data should produce no diagnostics
	if len(diagParams.Diagnostics) != 0 {
		t.Errorf("Expected 0 diagnostics for valid complex .sup data, got %d: %v",
			len(diagParams.Diagnostics), diagParams.Diagnostics)
	}
}

func TestInitializeWithNewCapabilities(t *testing.T) {
	h := NewTestHelper()

	params := InitializeParams{
		ProcessID: 1234,
		RootURI:   "file:///test",
	}

	response, err := h.ProcessRequest(1, "initialize", params)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Marshal result: %v", err)
	}

	var result InitializeResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("Unmarshal result: %v", err)
	}

	// Check hover capability
	if !result.Capabilities.HoverProvider {
		t.Error("Expected HoverProvider to be true")
	}

	// Check signature help capability
	if result.Capabilities.SignatureHelpProvider == nil {
		t.Error("Expected SignatureHelpProvider to be set")
	}

	// Check formatting capability
	if !result.Capabilities.DocumentFormattingProvider {
		t.Error("Expected DocumentFormattingProvider to be true")
	}
}
