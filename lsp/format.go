package main

import (
	"strings"
	"unicode"
)

// formatDocument formats a SuperSQL document
func formatDocument(text string, options FormattingOptions) string {
	// Tokenize and format
	tokens := tokenize(text)
	return formatTokens(tokens, options)
}

// Token types for formatting
type tokenType int

const (
	tokWhitespace tokenType = iota
	tokNewline
	tokComment
	tokString
	tokRegexp
	tokIdentifier
	tokNumber
	tokOperator
	tokPipe
	tokPunctuation
	tokKeyword
)

type token struct {
	typ   tokenType
	value string
}

// tokenize breaks the input into tokens
func tokenize(text string) []token {
	var tokens []token
	i := 0

	for i < len(text) {
		ch := text[i]

		// Newlines
		if ch == '\n' {
			tokens = append(tokens, token{tokNewline, "\n"})
			i++
			continue
		}

		// Whitespace (not newlines)
		if ch == ' ' || ch == '\t' || ch == '\r' {
			start := i
			for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\r') {
				i++
			}
			tokens = append(tokens, token{tokWhitespace, text[start:i]})
			continue
		}

		// Line comments (--)
		if i+1 < len(text) && ch == '-' && text[i+1] == '-' {
			start := i
			for i < len(text) && text[i] != '\n' {
				i++
			}
			tokens = append(tokens, token{tokComment, text[start:i]})
			continue
		}

		// Block comments (/* */)
		if i+1 < len(text) && ch == '/' && text[i+1] == '*' {
			start := i
			i += 2
			for i+1 < len(text) && !(text[i] == '*' && text[i+1] == '/') {
				i++
			}
			if i+1 < len(text) {
				i += 2
			}
			tokens = append(tokens, token{tokComment, text[start:i]})
			continue
		}

		// Strings (double-quoted, single-quoted, f-strings, raw strings)
		if ch == '"' || ch == '\'' || (ch == 'f' && i+1 < len(text) && (text[i+1] == '"' || text[i+1] == '\'')) ||
			(ch == 'r' && i+1 < len(text) && (text[i+1] == '"' || text[i+1] == '\'')) {
			start := i
			quote := ch
			if ch == 'f' || ch == 'r' {
				i++
				quote = text[i]
			}
			i++ // skip opening quote
			for i < len(text) && text[i] != byte(quote) {
				if text[i] == '\\' && i+1 < len(text) {
					i += 2
				} else {
					i++
				}
			}
			if i < len(text) {
				i++ // skip closing quote
			}
			tokens = append(tokens, token{tokString, text[start:i]})
			continue
		}

		// Regex literals
		if ch == '/' && (len(tokens) == 0 || canPrecedeRegex(tokens[len(tokens)-1])) {
			start := i
			i++ // skip opening /
			for i < len(text) && text[i] != '/' && text[i] != '\n' {
				if text[i] == '\\' && i+1 < len(text) {
					i += 2
				} else {
					i++
				}
			}
			if i < len(text) && text[i] == '/' {
				i++ // skip closing /
				tokens = append(tokens, token{tokRegexp, text[start:i]})
				continue
			}
			// Not a regex, treat as operator
			i = start
		}

		// Pipe operators
		if ch == '|' {
			if i+1 < len(text) && text[i+1] == '>' {
				tokens = append(tokens, token{tokPipe, "|>"})
				i += 2
				continue
			}
			if i+1 < len(text) && text[i+1] == '|' {
				// String concatenation operator
				tokens = append(tokens, token{tokOperator, "||"})
				i += 2
				continue
			}
			tokens = append(tokens, token{tokPipe, "|"})
			i++
			continue
		}

		// Multi-character operators
		if i+2 < len(text) && text[i:i+3] == "..." {
			tokens = append(tokens, token{tokOperator, "..."})
			i += 3
			continue
		}
		if i+1 < len(text) {
			twoChar := text[i : i+2]
			if twoChar == ":=" || twoChar == "::" || twoChar == "->" ||
				twoChar == "==" || twoChar == "!=" || twoChar == "<>" ||
				twoChar == "<=" || twoChar == ">=" || twoChar == "!~" {
				tokens = append(tokens, token{tokOperator, twoChar})
				i += 2
				continue
			}
		}

		// Single-character operators and punctuation
		if strings.ContainsRune("+-*/%<>=!~", rune(ch)) {
			tokens = append(tokens, token{tokOperator, string(ch)})
			i++
			continue
		}

		if strings.ContainsRune("()[]{},:;.?", rune(ch)) {
			tokens = append(tokens, token{tokPunctuation, string(ch)})
			i++
			continue
		}

		// Numbers
		if isDigit(ch) || (ch == '.' && i+1 < len(text) && isDigit(text[i+1])) {
			start := i
			// Handle hex
			if ch == '0' && i+1 < len(text) && (text[i+1] == 'x' || text[i+1] == 'X') {
				i += 2
				for i < len(text) && isHexDigit(text[i]) {
					i++
				}
			} else {
				for i < len(text) && (isDigit(text[i]) || text[i] == '.') {
					i++
				}
				// Handle exponent
				if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
					i++
					if i < len(text) && (text[i] == '+' || text[i] == '-') {
						i++
					}
					for i < len(text) && isDigit(text[i]) {
						i++
					}
				}
			}
			// Handle duration suffixes
			for i < len(text) && isLetter(text[i]) {
				i++
			}
			tokens = append(tokens, token{tokNumber, text[start:i]})
			continue
		}

		// Identifiers and keywords
		if isLetter(ch) || ch == '_' || ch == '`' {
			start := i
			if ch == '`' {
				// Backtick-quoted identifier
				i++
				for i < len(text) && text[i] != '`' {
					i++
				}
				if i < len(text) {
					i++
				}
			} else {
				for i < len(text) && (isLetter(text[i]) || isDigit(text[i]) || text[i] == '_') {
					i++
				}
			}
			word := text[start:i]
			if isKeyword(word) {
				tokens = append(tokens, token{tokKeyword, word})
			} else {
				tokens = append(tokens, token{tokIdentifier, word})
			}
			continue
		}

		// Unknown character - preserve it
		tokens = append(tokens, token{tokPunctuation, string(ch)})
		i++
	}

	return tokens
}

// formatTokens formats tokens into a string
func formatTokens(tokens []token, options FormattingOptions) string {
	var result strings.Builder
	indent := 0
	indentStr := "\t"
	if options.InsertSpaces {
		indentStr = strings.Repeat(" ", options.TabSize)
	}

	lineStart := true
	prevTok := token{}

	for i, tok := range tokens {
		switch tok.typ {
		case tokNewline:
			result.WriteString("\n")
			lineStart = true

		case tokWhitespace:
			// Normalize whitespace to single space (unless at line start, before pipe/newline, or after pipe)
			if !lineStart && i+1 < len(tokens) && tokens[i+1].typ != tokNewline &&
				tokens[i+1].typ != tokPipe && prevTok.typ != tokPipe {
				result.WriteString(" ")
			}

		case tokComment:
			if lineStart {
				result.WriteString(strings.Repeat(indentStr, indent))
			}
			result.WriteString(tok.value)
			lineStart = false

		case tokPipe:
			// Put pipe on new line with proper indentation
			if !lineStart {
				result.WriteString("\n")
			}
			result.WriteString(strings.Repeat(indentStr, indent))
			result.WriteString(tok.value)
			result.WriteString(" ")
			lineStart = false

		case tokPunctuation:
			if lineStart && tok.value != ")" && tok.value != "]" && tok.value != "}" {
				result.WriteString(strings.Repeat(indentStr, indent))
			}

			// Handle spacing around punctuation
			switch tok.value {
			case "(":
				result.WriteString(tok.value)
				indent++
			case ")":
				indent--
				if indent < 0 {
					indent = 0
				}
				result.WriteString(tok.value)
			case "{":
				result.WriteString(tok.value)
				indent++
			case "}":
				indent--
				if indent < 0 {
					indent = 0
				}
				if lineStart {
					result.WriteString(strings.Repeat(indentStr, indent))
				}
				result.WriteString(tok.value)
			case "[":
				result.WriteString(tok.value)
			case "]":
				result.WriteString(tok.value)
			case ",":
				result.WriteString(tok.value)
				// Add space after comma
				if i+1 < len(tokens) && tokens[i+1].typ != tokNewline {
					result.WriteString(" ")
				}
			case ":":
				// Space around colon in records
				if prevTok.typ == tokIdentifier || prevTok.typ == tokString {
					result.WriteString(": ")
				} else {
					result.WriteString(tok.value)
				}
			case ";":
				result.WriteString(tok.value)
			case ".":
				result.WriteString(tok.value)
			default:
				result.WriteString(tok.value)
			}
			lineStart = false

		case tokOperator:
			if lineStart {
				result.WriteString(strings.Repeat(indentStr, indent))
			}
			// Space around most operators
			switch tok.value {
			case ".", "...":
				result.WriteString(tok.value)
			case "::", "->":
				result.WriteString(tok.value)
			default:
				// Add space before if not at line start and prev wasn't space-producing
				if !lineStart && prevTok.typ != tokWhitespace && prevTok.typ != tokNewline &&
					prevTok.value != "(" && prevTok.value != "[" {
					result.WriteString(" ")
				}
				result.WriteString(tok.value)
				// Add space after
				if i+1 < len(tokens) && tokens[i+1].typ != tokNewline &&
					tokens[i+1].typ != tokWhitespace && tokens[i+1].value != ")" &&
					tokens[i+1].value != "]" && tokens[i+1].value != "," {
					result.WriteString(" ")
				}
			}
			lineStart = false

		case tokKeyword:
			if lineStart {
				result.WriteString(strings.Repeat(indentStr, indent))
			} else if needsSpaceBefore(prevTok) {
				result.WriteString(" ")
			}
			result.WriteString(tok.value)
			lineStart = false

		case tokIdentifier, tokNumber, tokString, tokRegexp:
			if lineStart {
				result.WriteString(strings.Repeat(indentStr, indent))
			} else if needsSpaceBefore(prevTok) {
				result.WriteString(" ")
			}
			result.WriteString(tok.value)
			lineStart = false
		}

		prevTok = tok
	}

	formatted := result.String()

	// Trim trailing whitespace from each line
	if options.TrimTrailingWhitespace {
		lines := strings.Split(formatted, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRightFunc(line, unicode.IsSpace)
		}
		formatted = strings.Join(lines, "\n")
	}

	// Handle final newline
	if options.TrimFinalNewlines {
		formatted = strings.TrimRight(formatted, "\n")
	}
	if options.InsertFinalNewline && !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	return formatted
}

func needsSpaceBefore(prev token) bool {
	switch prev.typ {
	case tokWhitespace, tokNewline, tokPipe:
		return false
	case tokPunctuation:
		return prev.value != "(" && prev.value != "[" && prev.value != "." && prev.value != ":"
	case tokOperator:
		return prev.value != "." && prev.value != "::" && prev.value != "->"
	default:
		return true
	}
}

func canPrecedeRegex(tok token) bool {
	switch tok.typ {
	case tokWhitespace, tokNewline, tokPipe, tokOperator:
		return true
	case tokPunctuation:
		return tok.value == "(" || tok.value == "[" || tok.value == "," || tok.value == ":"
	case tokKeyword:
		return true
	default:
		return false
	}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// SQL and SuperSQL keywords for formatting purposes
var formattingKeywords = map[string]bool{
	"select": true, "from": true, "where": true, "group": true, "by": true,
	"having": true, "order": true, "limit": true, "offset": true, "with": true,
	"join": true, "inner": true, "left": true, "right": true, "outer": true,
	"full": true, "cross": true, "anti": true, "on": true, "using": true,
	"and": true, "or": true, "not": true, "in": true, "like": true, "is": true,
	"between": true, "case": true, "when": true, "then": true, "else": true,
	"end": true, "as": true, "distinct": true, "all": true, "union": true,
	"const": true, "fn": true, "op": true, "type": true, "let": true,
	"true": true, "false": true, "null": true, "asc": true, "desc": true,
}

func isKeyword(word string) bool {
	return formattingKeywords[strings.ToLower(word)]
}
