package main

import (
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/brimdata/super"
	"github.com/brimdata/super/scode"
	"github.com/brimdata/super/sup"
)

// parseDataFileAndGetDiagnostics parses a SUP data file and returns diagnostics
func parseDataFileAndGetDiagnostics(text string) []Diagnostic {
	var diagnostics []Diagnostic

	reader := strings.NewReader(text)
	parser := sup.NewParser(reader)
	sctx := super.NewContext()
	analyzer := sup.NewAnalyzer()
	builder := scode.NewBuilder()

	for {
		ast, err := parser.ParseValue()
		if err != nil {
			diag := dataErrorToDiagnostic(text, err)
			diagnostics = append(diagnostics, diag)
			break
		}
		if ast == nil {
			// End of input
			break
		}
		// Convert the AST to a value to catch semantic errors
		val, err := analyzer.ConvertValue(sctx, ast)
		if err != nil {
			diag := dataErrorToDiagnostic(text, err)
			diagnostics = append(diagnostics, diag)
			// Continue parsing to find more errors
			continue
		}
		// Also try to build the value to catch any build errors
		_, err = sup.Build(builder, val)
		if err != nil {
			diag := dataErrorToDiagnostic(text, err)
			diagnostics = append(diagnostics, diag)
		}
	}

	return diagnostics
}

// dataErrorToDiagnostic converts a data parser error to an LSP diagnostic
func dataErrorToDiagnostic(text string, err error) Diagnostic {
	errStr := err.Error()

	// Try to extract position from error message
	line, col := extractDataErrorPosition(errStr)

	// Calculate range from position
	rng := positionToRange(text, line, col)

	return Diagnostic{
		Range:    rng,
		Severity: DiagnosticSeverityError,
		Source:   "superdb-lsp",
		Message:  cleanDataErrorMessage(errStr),
	}
}

// extractDataErrorPosition tries to extract line and column from data parser error message
func extractDataErrorPosition(errStr string) (line, col int) {
	// Default to start of document
	line = 0
	col = 0

	// Try various patterns used by the SUP parser
	patterns := []string{
		`parse error at line (\d+), column (\d+)`,
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

// cleanDataErrorMessage removes position info from error message for cleaner display
func cleanDataErrorMessage(errStr string) string {
	// Remove common position prefixes
	patterns := []string{
		`parse error at line \d+, column \d+: `,
		`parse error: `,
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

// parseDataValues parses a SUP data file and returns the parsed values
// This is used for formatting and other operations that need the parsed data
func parseDataValues(text string) ([]*super.Value, error) {
	var values []*super.Value

	reader := strings.NewReader(text)
	parser := sup.NewParser(reader)
	sctx := super.NewContext()
	analyzer := sup.NewAnalyzer()
	builder := scode.NewBuilder()

	for {
		ast, err := parser.ParseValue()
		if err != nil {
			if err == io.EOF {
				break
			}
			return values, err
		}
		if ast == nil {
			break
		}

		val, err := analyzer.ConvertValue(sctx, ast)
		if err != nil {
			return values, err
		}

		superVal, err := sup.Build(builder, val)
		if err != nil {
			return values, err
		}

		valueCopy := superVal
		values = append(values, &valueCopy)
	}

	return values, nil
}
