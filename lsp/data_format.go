package main

import (
	"strings"

	"github.com/brimdata/super"
	"github.com/brimdata/super/scode"
	"github.com/brimdata/super/sup"
)

// formatDataDocument formats a SUP data file
func formatDataDocument(text string, options FormattingOptions) string {
	// Parse the data values
	values, err := parseDataValuesForFormat(text)
	if err != nil {
		// If parsing fails, return the original text
		return text
	}

	if len(values) == 0 {
		return text
	}

	// Determine indentation
	indentSize := 0
	if options.InsertSpaces {
		indentSize = options.TabSize
	} else {
		indentSize = 4 // Use 4-space equivalent for tabs
	}

	// Format each value with the sup.Formatter
	formatter := sup.NewFormatter(indentSize, true, nil)

	var result strings.Builder
	for i, val := range values {
		if i > 0 {
			result.WriteString("\n")
		}
		formatted := formatter.FormatValue(*val)
		result.WriteString(formatted)
	}

	formatted := result.String()

	// Handle trailing whitespace
	if options.TrimTrailingWhitespace {
		lines := strings.Split(formatted, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " \t")
		}
		formatted = strings.Join(lines, "\n")
	}

	// Handle final newlines
	if options.TrimFinalNewlines {
		formatted = strings.TrimRight(formatted, "\n")
	}
	if options.InsertFinalNewline && !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	return formatted
}

// parseDataValuesForFormat parses a SUP data file and returns the parsed values
// This is specifically for formatting - it uses a fresh context and returns on first error
func parseDataValuesForFormat(text string) ([]*super.Value, error) {
	var values []*super.Value

	reader := strings.NewReader(text)
	parser := sup.NewParser(reader)
	sctx := super.NewContext()
	analyzer := sup.NewAnalyzer()
	builder := scode.NewBuilder()

	for {
		ast, err := parser.ParseValue()
		if err != nil {
			if len(values) > 0 {
				// We have some valid values, return those
				return values, nil
			}
			return nil, err
		}
		if ast == nil {
			break
		}

		val, err := analyzer.ConvertValue(sctx, ast)
		if err != nil {
			if len(values) > 0 {
				return values, nil
			}
			return nil, err
		}

		superVal, err := sup.Build(builder, val)
		if err != nil {
			if len(values) > 0 {
				return values, nil
			}
			return nil, err
		}

		valueCopy := superVal
		values = append(values, &valueCopy)
	}

	return values, nil
}

// isDataFile checks if a URI represents a .sup data file
func isDataFile(uri string) bool {
	return strings.HasSuffix(strings.ToLower(uri), ".sup")
}
