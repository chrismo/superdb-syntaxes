package main

import (
	"fmt"
	"strings"
)

// getHover returns hover information for the word at the given position
func getHover(text string, pos Position) *Hover {
	word := getWordAtPosition(text, pos)
	if word == "" {
		return nil
	}

	wordLower := strings.ToLower(word)

	// Check keywords
	for _, kw := range keywords {
		if strings.ToLower(kw.name) == wordLower {
			return &Hover{
				Contents: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: fmt.Sprintf("**%s** (keyword)\n\n%s", kw.name, kw.detail),
				},
			}
		}
	}

	// Check operators
	for _, op := range operators {
		if strings.ToLower(op.name) == wordLower {
			return &Hover{
				Contents: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: fmt.Sprintf("**%s** (operator)\n\n%s", op.name, op.detail),
				},
			}
		}
	}

	// Check functions
	for _, fn := range functions {
		if strings.ToLower(fn.name) == wordLower {
			sig := getFunctionSignature(fn.name)
			if sig != nil {
				return &Hover{
					Contents: MarkupContent{
						Kind:  MarkupKindMarkdown,
						Value: fmt.Sprintf("```spq\n%s\n```\n\n%s", sig.Label, fn.detail),
					},
				}
			}
			return &Hover{
				Contents: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: fmt.Sprintf("**%s** (function)\n\n%s", fn.name, fn.detail),
				},
			}
		}
	}

	// Check aggregates
	for _, agg := range aggregates {
		if strings.ToLower(agg.name) == wordLower {
			sig := getAggregateSignature(agg.name)
			if sig != nil {
				return &Hover{
					Contents: MarkupContent{
						Kind:  MarkupKindMarkdown,
						Value: fmt.Sprintf("```spq\n%s\n```\n\n%s", sig.Label, agg.detail),
					},
				}
			}
			return &Hover{
				Contents: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: fmt.Sprintf("**%s** (aggregate)\n\n%s", agg.name, agg.detail),
				},
			}
		}
	}

	// Check types
	for _, t := range types {
		if strings.ToLower(t.name) == wordLower {
			return &Hover{
				Contents: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: fmt.Sprintf("**%s** (type)\n\n%s", t.name, t.detail),
				},
			}
		}
	}

	return nil
}

// getWordAtPosition extracts the word at the given position
func getWordAtPosition(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return ""
	}

	line := lines[pos.Line]
	if pos.Character > len(line) {
		return ""
	}

	// Find word boundaries
	start := pos.Character
	end := pos.Character

	// Move start backward to find word start
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	// Move end forward to find word end
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}
