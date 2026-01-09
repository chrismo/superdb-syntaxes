package main

import (
	"regexp"
	"strings"
)

// MigrationDiagnostic represents a deprecated syntax diagnostic with a fix
type MigrationDiagnostic struct {
	Diagnostic Diagnostic
	Fix        *TextEdit // nil if no automatic fix available
}

// Migration represents a deprecated syntax pattern
type Migration struct {
	Code        string // Diagnostic code
	Pattern     *regexp.Regexp
	OldText     string // For display in message
	NewText     string // Replacement text (empty if no fix)
	Message     string
	Severity    int
	HasAutoFix  bool
	FixFunc     func(match string) string // Custom fix function
}

// Migrations for Phase 1: Simple Token Replacements
var migrations = []Migration{
	// Keyword renames
	{
		Code:       "deprecated-yield",
		Pattern:    regexp.MustCompile(`\byield\b`),
		OldText:    "yield",
		NewText:    "values",
		Message:    "'yield' is deprecated, use 'values'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
	},
	{
		Code:       "deprecated-func",
		Pattern:    regexp.MustCompile(`\bfunc\b`),
		OldText:    "func",
		NewText:    "fn",
		Message:    "'func' is deprecated, use 'fn'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
	},
	// Note: 'over' as an operator is still valid, this is for the deprecated usage
	// Skip 'over' for now as it requires semantic analysis to distinguish uses

	// Arrow operator
	{
		Code:       "deprecated-arrow",
		Pattern:    regexp.MustCompile(`=>`),
		OldText:    "=>",
		NewText:    "into",
		Message:    "'=>' is deprecated, use 'into'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
	},

	// Comment syntax - match // but not inside strings
	{
		Code:       "deprecated-comment-slash",
		Pattern:    regexp.MustCompile(`(^|[^:])//`), // Avoid matching :// in URLs
		OldText:    "//",
		NewText:    "--",
		Message:    "'//' comments are deprecated, use '--'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			// Preserve any character before the //
			if len(match) > 2 {
				return match[:len(match)-2] + "--"
			}
			return "--"
		},
	},

	// Function renames
	{
		Code:       "deprecated-parse-zson",
		Pattern:    regexp.MustCompile(`\bparse_zson\s*\(`),
		OldText:    "parse_zson",
		NewText:    "parse_sup",
		Message:    "'parse_zson' is deprecated, use 'parse_sup'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			return "parse_sup("
		},
	},

	// Phase 2: Implicit 'this' argument
	{
		Code:       "implicit-this-grep",
		Pattern:    regexp.MustCompile(`\bgrep\s*\(\s*(/[^/]*/|'[^']*'|"[^"]*")\s*\)`),
		OldText:    "grep(pattern)",
		NewText:    "grep(pattern, this)",
		Message:    "grep() requires explicit 'this' argument",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			// Extract the pattern and add this
			re := regexp.MustCompile(`\bgrep\s*\(\s*(/[^/]*/|'[^']*'|"[^"]*")\s*\)`)
			submatch := re.FindStringSubmatch(match)
			if len(submatch) > 1 {
				pattern := submatch[1]
				// Convert regex to string if needed
				if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
					// Convert /pattern/ to 'pattern'
					inner := pattern[1 : len(pattern)-1]
					return "grep('" + inner + "', this)"
				}
				return "grep(" + pattern + ", this)"
			}
			return match
		},
	},
	{
		Code:       "implicit-this-is",
		Pattern:    regexp.MustCompile(`\bis\s*\(\s*<[^>]+>\s*\)`),
		OldText:    "is(<type>)",
		NewText:    "is(this, <type>)",
		Message:    "is() requires explicit 'this' argument",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			// Extract the type and add this as first argument
			re := regexp.MustCompile(`\bis\s*\(\s*(<[^>]+>)\s*\)`)
			submatch := re.FindStringSubmatch(match)
			if len(submatch) > 1 {
				typeArg := submatch[1]
				return "is(this, " + typeArg + ")"
			}
			return match
		},
	},
	{
		Code:       "implicit-this-nest-dotted",
		Pattern:    regexp.MustCompile(`\bnest_dotted\s*\(\s*\)`),
		OldText:    "nest_dotted()",
		NewText:    "nest_dotted(this)",
		Message:    "nest_dotted() requires explicit 'this' argument",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			return "nest_dotted(this)"
		},
	},

	// Phase 2: Cast syntax
	{
		Code:       "deprecated-cast-time",
		Pattern:    regexp.MustCompile(`\btime\s*\(\s*('[^']*'|"[^"]*")\s*\)`),
		OldText:    "time('...')",
		NewText:    "'...'::time",
		Message:    "Function-style cast deprecated, use '::time'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			re := regexp.MustCompile(`\btime\s*\(\s*('[^']*'|"[^"]*")\s*\)`)
			submatch := re.FindStringSubmatch(match)
			if len(submatch) > 1 {
				return submatch[1] + "::time"
			}
			return match
		},
	},
	{
		Code:       "deprecated-cast-duration",
		Pattern:    regexp.MustCompile(`\bduration\s*\(\s*('[^']*'|"[^"]*")\s*\)`),
		OldText:    "duration('...')",
		NewText:    "'...'::duration",
		Message:    "Function-style cast deprecated, use '::duration'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			re := regexp.MustCompile(`\bduration\s*\(\s*('[^']*'|"[^"]*")\s*\)`)
			submatch := re.FindStringSubmatch(match)
			if len(submatch) > 1 {
				return submatch[1] + "::duration"
			}
			return match
		},
	},
	{
		Code:       "deprecated-cast-ip",
		Pattern:    regexp.MustCompile(`\bip\s*\(\s*('[^']*'|"[^"]*")\s*\)`),
		OldText:    "ip('...')",
		NewText:    "'...'::ip",
		Message:    "Function-style cast deprecated, use '::ip'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			re := regexp.MustCompile(`\bip\s*\(\s*('[^']*'|"[^"]*")\s*\)`)
			submatch := re.FindStringSubmatch(match)
			if len(submatch) > 1 {
				return submatch[1] + "::ip"
			}
			return match
		},
	},
	{
		Code:       "deprecated-cast-net",
		Pattern:    regexp.MustCompile(`\bnet\s*\(\s*('[^']*'|"[^"]*")\s*\)`),
		OldText:    "net('...')",
		NewText:    "'...'::net",
		Message:    "Function-style cast deprecated, use '::net'",
		Severity:   DiagnosticSeverityWarning,
		HasAutoFix: true,
		FixFunc: func(match string) string {
			re := regexp.MustCompile(`\bnet\s*\(\s*('[^']*'|"[^"]*")\s*\)`)
			submatch := re.FindStringSubmatch(match)
			if len(submatch) > 1 {
				return submatch[1] + "::net"
			}
			return match
		},
	},

	// Phase 4: Removed functions (no auto-fix)
	{
		Code:       "removed-crop",
		Pattern:    regexp.MustCompile(`\bcrop\s*\(`),
		OldText:    "crop()",
		Message:    "'crop()' was removed, use explicit casting",
		Severity:   DiagnosticSeverityError,
		HasAutoFix: false,
	},
	{
		Code:       "removed-fill",
		Pattern:    regexp.MustCompile(`\bfill\s*\(`),
		OldText:    "fill()",
		Message:    "'fill()' was removed, use explicit casting",
		Severity:   DiagnosticSeverityError,
		HasAutoFix: false,
	},
	{
		Code:       "removed-fit",
		Pattern:    regexp.MustCompile(`\bfit\s*\(`),
		OldText:    "fit()",
		Message:    "'fit()' was removed, use explicit casting",
		Severity:   DiagnosticSeverityError,
		HasAutoFix: false,
	},
	{
		Code:       "removed-order",
		Pattern:    regexp.MustCompile(`\border\s*\(`),
		OldText:    "order()",
		Message:    "'order()' was removed, use explicit casting",
		Severity:   DiagnosticSeverityError,
		HasAutoFix: false,
	},
	{
		Code:       "removed-shape",
		Pattern:    regexp.MustCompile(`\bshape\s*\(`),
		OldText:    "shape()",
		Message:    "'shape()' was removed, use explicit casting",
		Severity:   DiagnosticSeverityError,
		HasAutoFix: false,
	},
}

// getMigrationDiagnostics scans text for deprecated syntax patterns
func getMigrationDiagnostics(text string) []MigrationDiagnostic {
	var diagnostics []MigrationDiagnostic
	lines := strings.Split(text, "\n")

	for lineNum, line := range lines {
		// Skip lines that are already using -- comments
		// to avoid false positives on comment content
		commentIdx := strings.Index(line, "--")

		for _, m := range migrations {
			matches := m.Pattern.FindAllStringIndex(line, -1)
			for _, match := range matches {
				startCol := match[0]
				endCol := match[1]

				// Skip matches inside -- comments
				if commentIdx >= 0 && startCol > commentIdx {
					continue
				}

				// Skip // detection if it's part of a URL (has : before it)
				if m.Code == "deprecated-comment-slash" {
					matchStr := line[startCol:endCol]
					// The pattern captures optional char before //, check if it's :
					if strings.Contains(matchStr, "://") {
						continue
					}
					// Adjust range to only highlight the //
					if len(matchStr) > 2 && matchStr[len(matchStr)-2:] == "//" {
						startCol = endCol - 2
					}
				}

				diag := MigrationDiagnostic{
					Diagnostic: Diagnostic{
						Range: Range{
							Start: Position{Line: lineNum, Character: startCol},
							End:   Position{Line: lineNum, Character: endCol},
						},
						Severity: m.Severity,
						Code:     m.Code,
						Source:   "superdb-lsp",
						Message:  m.Message,
					},
				}

				if m.HasAutoFix {
					matchedText := line[match[0]:match[1]]
					var newText string
					if m.FixFunc != nil {
						newText = m.FixFunc(matchedText)
					} else {
						newText = m.NewText
					}

					diag.Fix = &TextEdit{
						Range: Range{
							Start: Position{Line: lineNum, Character: match[0]},
							End:   Position{Line: lineNum, Character: match[1]},
						},
						NewText: newText,
					}
				}

				diagnostics = append(diagnostics, diag)
			}
		}
	}

	return diagnostics
}

// getCodeActionsForDiagnostics generates code actions for migration diagnostics
func getCodeActionsForDiagnostics(uri string, text string, requestedDiags []Diagnostic) []CodeAction {
	var actions []CodeAction

	// Get all migration diagnostics for this document
	migrationDiags := getMigrationDiagnostics(text)

	// Build a map of fixable diagnostics by code+range
	fixableDiags := make(map[string]MigrationDiagnostic)
	for _, md := range migrationDiags {
		if md.Fix != nil {
			key := diagnosticKey(md.Diagnostic)
			fixableDiags[key] = md
		}
	}

	// Create individual quick-fix actions for requested diagnostics
	for _, reqDiag := range requestedDiags {
		key := diagnosticKey(reqDiag)
		if md, ok := fixableDiags[key]; ok {
			action := CodeAction{
				Title:       "Replace with '" + md.Fix.NewText + "'",
				Kind:        CodeActionKindQuickFix,
				Diagnostics: []Diagnostic{md.Diagnostic},
				IsPreferred: true,
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						uri: {*md.Fix},
					},
				},
			}
			actions = append(actions, action)
		}
	}

	// Create "Fix all migration issues" action if there are multiple fixes
	if len(fixableDiags) > 1 {
		var allEdits []TextEdit
		var allDiags []Diagnostic
		for _, md := range fixableDiags {
			allEdits = append(allEdits, *md.Fix)
			allDiags = append(allDiags, md.Diagnostic)
		}

		// Sort edits by position (reverse order for safe application)
		sortEditsReverse(allEdits)

		fixAllAction := CodeAction{
			Title:       "Fix all deprecated syntax",
			Kind:        CodeActionKindSourceFixAll,
			Diagnostics: allDiags,
			Edit: &WorkspaceEdit{
				Changes: map[string][]TextEdit{
					uri: allEdits,
				},
			},
		}
		actions = append(actions, fixAllAction)
	}

	return actions
}

// diagnosticKey creates a unique key for a diagnostic
func diagnosticKey(d Diagnostic) string {
	return d.Code + ":" +
		string(rune(d.Range.Start.Line)) + ":" +
		string(rune(d.Range.Start.Character)) + ":" +
		string(rune(d.Range.End.Line)) + ":" +
		string(rune(d.Range.End.Character))
}

// sortEditsReverse sorts edits in reverse document order (bottom to top, right to left)
// This ensures edits don't invalidate each other's positions
func sortEditsReverse(edits []TextEdit) {
	for i := 0; i < len(edits)-1; i++ {
		for j := i + 1; j < len(edits); j++ {
			if comparePositions(edits[i].Range.Start, edits[j].Range.Start) < 0 {
				edits[i], edits[j] = edits[j], edits[i]
			}
		}
	}
}

// comparePositions compares two positions, returning -1, 0, or 1
func comparePositions(a, b Position) int {
	if a.Line < b.Line {
		return -1
	}
	if a.Line > b.Line {
		return 1
	}
	if a.Character < b.Character {
		return -1
	}
	if a.Character > b.Character {
		return 1
	}
	return 0
}
