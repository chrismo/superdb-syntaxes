package main

import (
	"encoding/json"
	"log"
)

// handleInitialize processes the initialize request
func (s *Server) handleInitialize(msg RPCMessage) (interface{}, error) {
	var params InitializeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	log.Printf("Initialize: processId=%d, rootUri=%s", params.ProcessID, params.RootURI)

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			// Full document sync - client sends entire document on change
			TextDocumentSync: 1,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", "|", "(", ":", "="},
				ResolveProvider:   false,
			},
			HoverProvider: true,
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters:   []string{"(", ","},
				RetriggerCharacters: []string{","},
			},
			DocumentFormattingProvider: true,
		},
		ServerInfo: &ServerInfo{
			Name:    "superdb-lsp",
			Version: Version,
		},
	}

	return RPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

// handleShutdown processes the shutdown request
func (s *Server) handleShutdown(msg RPCMessage) (interface{}, error) {
	log.Println("Shutdown requested")
	s.shutdown = true

	return RPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  nil,
	}, nil
}

// handleDidOpen processes textDocument/didOpen notifications
func (s *Server) handleDidOpen(msg RPCMessage) (interface{}, error) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	text := params.TextDocument.Text

	log.Printf("Document opened: %s (lang=%s, version=%d)",
		uri, params.TextDocument.LanguageID, params.TextDocument.Version)

	s.documents[uri] = text

	// Parse and publish diagnostics
	return s.publishDiagnostics(uri, text, params.TextDocument.Version)
}

// handleDidChange processes textDocument/didChange notifications
func (s *Server) handleDidChange(msg RPCMessage) (interface{}, error) {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI

	// With TextDocumentSync=1 (Full), we get the full document content
	if len(params.ContentChanges) > 0 {
		text := params.ContentChanges[len(params.ContentChanges)-1].Text
		s.documents[uri] = text

		log.Printf("Document changed: %s (version=%d)", uri, params.TextDocument.Version)

		// Parse and publish diagnostics
		return s.publishDiagnostics(uri, text, params.TextDocument.Version)
	}

	return nil, nil
}

// handleDidClose processes textDocument/didClose notifications
func (s *Server) handleDidClose(msg RPCMessage) (interface{}, error) {
	var params DidCloseTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	delete(s.documents, uri)

	log.Printf("Document closed: %s", uri)

	return nil, nil
}

// handleCompletion processes textDocument/completion requests
func (s *Server) handleCompletion(msg RPCMessage) (interface{}, error) {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	text, ok := s.documents[uri]
	if !ok {
		log.Printf("Document not found: %s", uri)
		return RPCMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  CompletionList{Items: []CompletionItem{}},
		}, nil
	}

	log.Printf("Completion request: %s at line=%d, char=%d",
		uri, params.Position.Line, params.Position.Character)

	items := getCompletions(text, params.Position)

	return RPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  CompletionList{Items: items},
	}, nil
}

// handleHover processes textDocument/hover requests
func (s *Server) handleHover(msg RPCMessage) (interface{}, error) {
	var params HoverParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	text, ok := s.documents[uri]
	if !ok {
		log.Printf("Document not found: %s", uri)
		return RPCMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  nil,
		}, nil
	}

	log.Printf("Hover request: %s at line=%d, char=%d",
		uri, params.Position.Line, params.Position.Character)

	hover := getHover(text, params.Position)

	return RPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  hover,
	}, nil
}

// handleSignatureHelp processes textDocument/signatureHelp requests
func (s *Server) handleSignatureHelp(msg RPCMessage) (interface{}, error) {
	var params SignatureHelpParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	text, ok := s.documents[uri]
	if !ok {
		log.Printf("Document not found: %s", uri)
		return RPCMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  nil,
		}, nil
	}

	log.Printf("Signature help request: %s at line=%d, char=%d",
		uri, params.Position.Line, params.Position.Character)

	sigHelp := getSignatureHelp(text, params.Position)

	return RPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  sigHelp,
	}, nil
}

// handleFormatting processes textDocument/formatting requests
func (s *Server) handleFormatting(msg RPCMessage) (interface{}, error) {
	var params DocumentFormattingParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	uri := params.TextDocument.URI
	text, ok := s.documents[uri]
	if !ok {
		log.Printf("Document not found: %s", uri)
		return RPCMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  []TextEdit{},
		}, nil
	}

	log.Printf("Formatting request: %s (tabSize=%d, insertSpaces=%v)",
		uri, params.Options.TabSize, params.Options.InsertSpaces)

	formatted := formatDocument(text, params.Options)

	// If no changes, return empty array
	if formatted == text {
		return RPCMessage{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  []TextEdit{},
		}, nil
	}

	// Return a single edit that replaces the entire document
	lines := len(splitLines(text))
	lastLineLen := 0
	if lines > 0 {
		lastLineLen = len(getLastLine(text))
	}

	edit := TextEdit{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: lines, Character: lastLineLen},
		},
		NewText: formatted,
	}

	return RPCMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  []TextEdit{edit},
	}, nil
}

// splitLines splits text into lines
func splitLines(text string) []string {
	if text == "" {
		return []string{}
	}
	lines := []string{}
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

// getLastLine returns the last line of text
func getLastLine(text string) string {
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == '\n' {
			return text[i+1:]
		}
	}
	return text
}
