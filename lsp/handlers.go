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
		},
		ServerInfo: &ServerInfo{
			Name:    "superdb-lsp",
			Version: "0.1.0",
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
