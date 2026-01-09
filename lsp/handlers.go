package main

import (
	"encoding/json"
	"log"
)

// response creates an RPCMessage response with the given ID and result
func response(id interface{}, result interface{}) (interface{}, error) {
	return RPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}, nil
}

// handleInitialize processes the initialize request
func (s *Server) handleInitialize(msg RPCMessage) (interface{}, error) {
	var params InitializeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	log.Printf("Initialize: processId=%d, rootUri=%s", params.ProcessID, params.RootURI)

	return response(msg.ID, InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full document sync
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
			CodeActionProvider: &CodeActionOptions{
				CodeActionKinds: []string{
					CodeActionKindQuickFix,
					CodeActionKindSourceFixAll,
				},
			},
		},
		ServerInfo: &ServerInfo{
			Name:    "superdb-lsp",
			Version: Version,
		},
	})
}

// handleShutdown processes the shutdown request
func (s *Server) handleShutdown(msg RPCMessage) (interface{}, error) {
	log.Println("Shutdown requested")
	s.shutdown = true
	return response(msg.ID, nil)
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

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		log.Printf("Document not found: %s", params.TextDocument.URI)
		return response(msg.ID, CompletionList{Items: []CompletionItem{}})
	}

	log.Printf("Completion request: %s at line=%d, char=%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	return response(msg.ID, CompletionList{Items: getCompletions(text, params.Position)})
}

// handleHover processes textDocument/hover requests
func (s *Server) handleHover(msg RPCMessage) (interface{}, error) {
	var params HoverParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		log.Printf("Document not found: %s", params.TextDocument.URI)
		return response(msg.ID, nil)
	}

	log.Printf("Hover request: %s at line=%d, char=%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	return response(msg.ID, getHover(text, params.Position))
}

// handleSignatureHelp processes textDocument/signatureHelp requests
func (s *Server) handleSignatureHelp(msg RPCMessage) (interface{}, error) {
	var params SignatureHelpParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		log.Printf("Document not found: %s", params.TextDocument.URI)
		return response(msg.ID, nil)
	}

	log.Printf("Signature help request: %s at line=%d, char=%d",
		params.TextDocument.URI, params.Position.Line, params.Position.Character)

	return response(msg.ID, getSignatureHelp(text, params.Position))
}

// handleFormatting processes textDocument/formatting requests
func (s *Server) handleFormatting(msg RPCMessage) (interface{}, error) {
	var params DocumentFormattingParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		log.Printf("Document not found: %s", params.TextDocument.URI)
		return response(msg.ID, []TextEdit{})
	}

	log.Printf("Formatting request: %s (tabSize=%d, insertSpaces=%v)",
		params.TextDocument.URI, params.Options.TabSize, params.Options.InsertSpaces)

	var formatted string
	if isDataFile(params.TextDocument.URI) {
		// Format as SUP data file
		formatted = formatDataDocument(text, params.Options)
	} else {
		// Format as SuperSQL query
		formatted = formatDocument(text, params.Options)
	}

	// If no changes, return empty array
	if formatted == text {
		return response(msg.ID, []TextEdit{})
	}

	// Return a single edit that replaces the entire document
	lines := splitLines(text)
	lastLineLen := 0
	if len(lines) > 0 {
		lastLineLen = len(lines[len(lines)-1])
	}

	return response(msg.ID, []TextEdit{{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: len(lines) - 1, Character: lastLineLen},
		},
		NewText: formatted,
	}})
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

// handleCodeAction processes textDocument/codeAction requests
func (s *Server) handleCodeAction(msg RPCMessage) (interface{}, error) {
	var params CodeActionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	text, ok := s.documents[params.TextDocument.URI]
	if !ok {
		log.Printf("Document not found: %s", params.TextDocument.URI)
		return response(msg.ID, []CodeAction{})
	}

	log.Printf("Code action request: %s at line=%d-%d",
		params.TextDocument.URI,
		params.Range.Start.Line,
		params.Range.End.Line)

	// Get code actions for the diagnostics in context
	actions := getCodeActionsForDiagnostics(params.TextDocument.URI, text, params.Context.Diagnostics)

	return response(msg.ID, actions)
}
