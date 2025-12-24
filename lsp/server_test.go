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
		"from test | yield {a: 1}",
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

func TestKeywordCount(t *testing.T) {
	// Verify we have a reasonable number of keywords
	if len(keywords) < 40 {
		t.Errorf("Expected at least 40 keywords, got %d", len(keywords))
	}
}

func TestOperatorCount(t *testing.T) {
	// Verify we have a reasonable number of operators
	if len(operators) < 25 {
		t.Errorf("Expected at least 25 operators, got %d", len(operators))
	}
}

func TestFunctionCount(t *testing.T) {
	// Verify we have a reasonable number of functions
	if len(functions) < 50 {
		t.Errorf("Expected at least 50 functions, got %d", len(functions))
	}
}

func TestTypeCount(t *testing.T) {
	// Verify we have a reasonable number of types
	if len(types) < 35 {
		t.Errorf("Expected at least 35 types, got %d", len(types))
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
