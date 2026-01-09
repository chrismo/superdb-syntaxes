package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/brimdata/super/compiler/parser"
)

// publishDiagnostics parses the document and publishes diagnostics
func (s *Server) publishDiagnostics(uri, text string, version int) (interface{}, error) {
	var diagnostics []Diagnostic
	if isDataFile(uri) {
		// Parse as SUP data file
		diagnostics = parseDataFileAndGetDiagnostics(text)
	} else {
		// Parse as SuperSQL query
		diagnostics = parseAndGetDiagnostics(text)
	}

	log.Printf("Publishing %d diagnostics for %s", len(diagnostics), uri)

	params := PublishDiagnosticsParams{
		URI:         uri,
		Version:     version,
		Diagnostics: diagnostics,
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// Return a notification (no ID, no response expected)
	return RPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params:  paramsBytes,
	}, nil
}

// parseAndGetDiagnostics parses SuperSQL code and returns diagnostics
func parseAndGetDiagnostics(text string) []Diagnostic {
	var diagnostics []Diagnostic

	// Parse using the brimdata/super compiler parser
	_, err := parser.ParseQuery(text)
	if err != nil {
		diag := errorToDiagnostic(text, err)
		diagnostics = append(diagnostics, diag)
	}

	// Add migration diagnostics for deprecated syntax
	migrationDiags := getMigrationDiagnostics(text)
	for _, md := range migrationDiags {
		diagnostics = append(diagnostics, md.Diagnostic)
	}

	return diagnostics
}

// errorToDiagnostic converts a parser error to an LSP diagnostic
func errorToDiagnostic(text string, err error) Diagnostic {
	errStr := err.Error()

	// Try to extract position from error message
	// Parser errors typically look like: "error parsing at line X, column Y: message"
	line, col := extractPosition(errStr)

	// Calculate range from position
	rng := positionToRange(text, line, col)

	return Diagnostic{
		Range:    rng,
		Severity: DiagnosticSeverityError,
		Source:   "superdb-lsp",
		Message:  cleanErrorMessage(errStr),
	}
}

// extractPosition tries to extract line and column from error message
func extractPosition(errStr string) (line, col int) {
	// Default to start of document
	line = 0
	col = 0

	// Try various patterns used by the parser
	patterns := []string{
		`line (\d+), column (\d+)`,
		`line (\d+):(\d+)`,
		`(\d+):(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(errStr)
		if len(matches) >= 3 {
			if l, err := strconv.Atoi(matches[1]); err == nil {
				line = l - 1 // Convert to 0-based
				if line < 0 {
					line = 0
				}
			}
			if c, err := strconv.Atoi(matches[2]); err == nil {
				col = c - 1 // Convert to 0-based
				if col < 0 {
					col = 0
				}
			}
			break
		}
	}

	return line, col
}

// positionToRange creates a range from a position, highlighting the token
func positionToRange(text string, line, col int) Range {
	lines := strings.Split(text, "\n")

	// Ensure line is within bounds
	if line >= len(lines) {
		line = len(lines) - 1
	}
	if line < 0 {
		line = 0
	}

	lineText := ""
	if line < len(lines) {
		lineText = lines[line]
	}

	// Ensure col is within bounds
	if col >= len(lineText) {
		col = len(lineText)
	}
	if col < 0 {
		col = 0
	}

	// Find the end of the current token
	endCol := col
	for endCol < len(lineText) && !isWhitespace(lineText[endCol]) {
		endCol++
	}

	// If we're at the end, highlight at least one character
	if endCol == col {
		endCol = col + 1
	}

	return Range{
		Start: Position{Line: line, Character: col},
		End:   Position{Line: line, Character: endCol},
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// cleanErrorMessage removes position info from error message for cleaner display
func cleanErrorMessage(errStr string) string {
	// Remove common position prefixes
	patterns := []string{
		`error parsing at line \d+, column \d+: `,
		`line \d+:\d+: `,
		`\d+:\d+: `,
	}

	result := errStr
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "")
	}

	return strings.TrimSpace(result)
}
