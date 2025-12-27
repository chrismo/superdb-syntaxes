package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// FormatTestCase represents a single formatting test case
type FormatTestCase struct {
	Name     string           `toml:"name"`
	Skip     bool             `toml:"skip,omitempty"`
	Options  FormatOptions    `toml:"options"`
	Input    string           `toml:"input"`
	Expected string           `toml:"expected"`
}

// FormatOptions mirrors FormattingOptions for TOML parsing
type FormatOptions struct {
	TabSize                int  `toml:"tabSize"`
	InsertSpaces           bool `toml:"insertSpaces"`
	TrimTrailingWhitespace bool `toml:"trimTrailingWhitespace"`
	InsertFinalNewline     bool `toml:"insertFinalNewline"`
	TrimFinalNewlines      bool `toml:"trimFinalNewlines"`
}

func TestFormatGolden(t *testing.T) {
	files, err := filepath.Glob("testdata/format/*.toml")
	if err != nil {
		t.Fatalf("failed to glob test files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("no golden test files found in testdata/format/")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read %s: %v", file, err)
			}

			var tc FormatTestCase
			if err := toml.Unmarshal(data, &tc); err != nil {
				t.Fatalf("failed to parse %s: %v", file, err)
			}

			if tc.Skip {
				t.Skip("test marked as skip")
			}

			opts := FormattingOptions{
				TabSize:                tc.Options.TabSize,
				InsertSpaces:           tc.Options.InsertSpaces,
				TrimTrailingWhitespace: tc.Options.TrimTrailingWhitespace,
				InsertFinalNewline:     tc.Options.InsertFinalNewline,
				TrimFinalNewlines:      tc.Options.TrimFinalNewlines,
			}

			// Default tabSize if not specified
			if opts.TabSize == 0 {
				opts.TabSize = 2
			}

			got := formatDocument(tc.Input, opts)

			if got != tc.Expected {
				t.Errorf("formatting mismatch for %q\n\nInput:\n%s\n\nExpected:\n%s\n\nGot:\n%s\n\nDiff:\nexpected: %q\ngot:      %q",
					tc.Name,
					indent(tc.Input),
					indent(tc.Expected),
					indent(got),
					tc.Expected,
					got,
				)
			}
		})
	}
}

// indent adds a prefix to each line for readable test output
func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "  | " + line
	}
	return strings.Join(lines, "\n")
}
