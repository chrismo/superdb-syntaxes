package main

import (
	"strings"
)

// FunctionSig holds function signature information
type FunctionSig struct {
	Label      string
	Doc        string
	Parameters []ParamInfo
}

// ParamInfo holds parameter information
type ParamInfo struct {
	Name string
	Doc  string
}

// Function signatures for built-in functions
var functionSignatures = map[string]*FunctionSig{
	"abs":            {Label: "abs(value: number) -> number", Doc: "Returns the absolute value", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric value"}}},
	"base64":         {Label: "base64(value: bytes|string) -> string", Doc: "Encode/decode base64", Parameters: []ParamInfo{{Name: "value", Doc: "Value to encode/decode"}}},
	"bucket":         {Label: "bucket(value: number, size: number) -> number", Doc: "Bucket values into ranges", Parameters: []ParamInfo{{Name: "value", Doc: "Value to bucket"}, {Name: "size", Doc: "Bucket size"}}},
	"cast":           {Label: "cast(value: any, type: type) -> any", Doc: "Cast value to type", Parameters: []ParamInfo{{Name: "value", Doc: "Value to cast"}, {Name: "type", Doc: "Target type"}}},
	"ceil":           {Label: "ceil(value: number) -> number", Doc: "Round up to nearest integer", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric value"}}},
	"cidr_match":     {Label: "cidr_match(network: net, ip: ip) -> bool", Doc: "Check if IP matches CIDR", Parameters: []ParamInfo{{Name: "network", Doc: "CIDR network"}, {Name: "ip", Doc: "IP address to check"}}},
	"coalesce":       {Label: "coalesce(value: any, ...) -> any", Doc: "Return first non-null value", Parameters: []ParamInfo{{Name: "value", Doc: "Values to check"}}},
	"compare":        {Label: "compare(a: any, b: any) -> int64", Doc: "Compare two values (-1, 0, 1)", Parameters: []ParamInfo{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Second value"}}},
	"date_part":      {Label: "date_part(part: string, time: time) -> int64", Doc: "Extract part from timestamp", Parameters: []ParamInfo{{Name: "part", Doc: "Part name (year, month, day, hour, minute, second)"}, {Name: "time", Doc: "Timestamp value"}}},
	"error":          {Label: "error(message: string) -> error", Doc: "Create error value", Parameters: []ParamInfo{{Name: "message", Doc: "Error message"}}},
	"fields":         {Label: "fields(record: record) -> [string]", Doc: "Get record field names", Parameters: []ParamInfo{{Name: "record", Doc: "Record value"}}},
	"flatten":        {Label: "flatten(record: record) -> record", Doc: "Flatten nested records", Parameters: []ParamInfo{{Name: "record", Doc: "Record to flatten"}}},
	"floor":          {Label: "floor(value: number) -> number", Doc: "Round down to nearest integer", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric value"}}},
	"grep":           {Label: "grep(pattern: string|regexp, value: any) -> bool", Doc: "Search for pattern", Parameters: []ParamInfo{{Name: "pattern", Doc: "Search pattern"}, {Name: "value", Doc: "Value to search"}}},
	"grok":           {Label: "grok(pattern: string, value: string) -> record", Doc: "Parse with grok pattern", Parameters: []ParamInfo{{Name: "pattern", Doc: "Grok pattern"}, {Name: "value", Doc: "String to parse"}}},
	"has":            {Label: "has(record: record, field: string) -> bool", Doc: "Check if field exists", Parameters: []ParamInfo{{Name: "record", Doc: "Record to check"}, {Name: "field", Doc: "Field name"}}},
	"has_error":      {Label: "has_error(value: any) -> bool", Doc: "Check for nested error", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}}},
	"hex":            {Label: "hex(value: bytes|string) -> string", Doc: "Convert to hexadecimal", Parameters: []ParamInfo{{Name: "value", Doc: "Value to convert"}}},
	"is":             {Label: "is(value: any, type: type) -> bool", Doc: "Check if value is type", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}, {Name: "type", Doc: "Type to check against"}}},
	"is_error":       {Label: "is_error(value: any) -> bool", Doc: "Check if value is error", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}}},
	"join":           {Label: "join(array: [string], sep: string) -> string", Doc: "Join strings with separator", Parameters: []ParamInfo{{Name: "array", Doc: "Array of strings"}, {Name: "sep", Doc: "Separator"}}},
	"kind":           {Label: "kind(value: any) -> string", Doc: "Get value kind", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}}},
	"ksuid":          {Label: "ksuid() -> string", Doc: "Generate KSUID", Parameters: []ParamInfo{}},
	"len":            {Label: "len(value: string|bytes|array) -> int64", Doc: "Get length", Parameters: []ParamInfo{{Name: "value", Doc: "Value to measure"}}},
	"length":         {Label: "length(value: string|bytes|array) -> int64", Doc: "Get length (alias)", Parameters: []ParamInfo{{Name: "value", Doc: "Value to measure"}}},
	"levenshtein":    {Label: "levenshtein(a: string, b: string) -> int64", Doc: "Levenshtein edit distance", Parameters: []ParamInfo{{Name: "a", Doc: "First string"}, {Name: "b", Doc: "Second string"}}},
	"log":            {Label: "log(value: number, base?: number) -> float64", Doc: "Logarithm", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric value"}, {Name: "base", Doc: "Log base (default: e)"}}},
	"lower":          {Label: "lower(value: string) -> string", Doc: "Convert to lowercase", Parameters: []ParamInfo{{Name: "value", Doc: "String to convert"}}},
	"max":            {Label: "max(a: number, b: number) -> number", Doc: "Maximum of two values", Parameters: []ParamInfo{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Second value"}}},
	"min":            {Label: "min(a: number, b: number) -> number", Doc: "Minimum of two values", Parameters: []ParamInfo{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Second value"}}},
	"missing":        {Label: "missing(type?: type) -> missing", Doc: "Create missing value", Parameters: []ParamInfo{{Name: "type", Doc: "Optional type"}}},
	"nameof":         {Label: "nameof(value: any) -> string", Doc: "Get type name", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}}},
	"nest_dotted":    {Label: "nest_dotted(record: record) -> record", Doc: "Nest dotted field names", Parameters: []ParamInfo{{Name: "record", Doc: "Record with dotted names"}}},
	"network_of":     {Label: "network_of(ip: ip, mask: net) -> net", Doc: "Get network from IP", Parameters: []ParamInfo{{Name: "ip", Doc: "IP address"}, {Name: "mask", Doc: "Network mask"}}},
	"now":            {Label: "now() -> time", Doc: "Current timestamp", Parameters: []ParamInfo{}},
	"nullif":         {Label: "nullif(a: any, b: any) -> any", Doc: "Return null if equal", Parameters: []ParamInfo{{Name: "a", Doc: "First value"}, {Name: "b", Doc: "Value to compare"}}},
	"parse_sup":      {Label: "parse_sup(value: string) -> any", Doc: "Parse Super format", Parameters: []ParamInfo{{Name: "value", Doc: "String to parse"}}},
	"parse_uri":      {Label: "parse_uri(uri: string) -> record", Doc: "Parse URI string", Parameters: []ParamInfo{{Name: "uri", Doc: "URI to parse"}}},
	"position":       {Label: "position(substr: string, str: string) -> int64", Doc: "Find substring position", Parameters: []ParamInfo{{Name: "substr", Doc: "Substring to find"}, {Name: "str", Doc: "String to search"}}},
	"pow":            {Label: "pow(base: number, exp: number) -> number", Doc: "Power function", Parameters: []ParamInfo{{Name: "base", Doc: "Base value"}, {Name: "exp", Doc: "Exponent"}}},
	"quiet":          {Label: "quiet(value: any) -> any", Doc: "Suppress errors", Parameters: []ParamInfo{{Name: "value", Doc: "Value to quiet"}}},
	"regexp":         {Label: "regexp(pattern: string, value: string) -> bool", Doc: "Regex match", Parameters: []ParamInfo{{Name: "pattern", Doc: "Regex pattern"}, {Name: "value", Doc: "String to match"}}},
	"regexp_replace": {Label: "regexp_replace(value: string, pattern: string, replacement: string) -> string", Doc: "Regex replacement", Parameters: []ParamInfo{{Name: "value", Doc: "Input string"}, {Name: "pattern", Doc: "Regex pattern"}, {Name: "replacement", Doc: "Replacement string"}}},
	"replace":        {Label: "replace(value: string, old: string, new: string) -> string", Doc: "String replacement", Parameters: []ParamInfo{{Name: "value", Doc: "Input string"}, {Name: "old", Doc: "String to replace"}, {Name: "new", Doc: "Replacement string"}}},
	"round":          {Label: "round(value: number, precision?: int64) -> number", Doc: "Round to precision", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric value"}, {Name: "precision", Doc: "Decimal places (default: 0)"}}},
	"split":          {Label: "split(value: string, sep: string) -> [string]", Doc: "Split string", Parameters: []ParamInfo{{Name: "value", Doc: "String to split"}, {Name: "sep", Doc: "Separator"}}},
	"sqrt":           {Label: "sqrt(value: number) -> float64", Doc: "Square root", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric value"}}},
	"strftime":       {Label: "strftime(format: string, time: time) -> string", Doc: "Format time as string", Parameters: []ParamInfo{{Name: "format", Doc: "Format string"}, {Name: "time", Doc: "Timestamp value"}}},
	"trim":           {Label: "trim(value: string) -> string", Doc: "Trim whitespace", Parameters: []ParamInfo{{Name: "value", Doc: "String to trim"}}},
	"typename":       {Label: "typename(value: any) -> string", Doc: "Get type name", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}}},
	"typeof":         {Label: "typeof(value: any) -> type", Doc: "Get type of value", Parameters: []ParamInfo{{Name: "value", Doc: "Value to check"}}},
	"under":          {Label: "under(value: any) -> any", Doc: "Get underlying value", Parameters: []ParamInfo{{Name: "value", Doc: "Value to unwrap"}}},
	"unflatten":      {Label: "unflatten(record: record) -> record", Doc: "Unflatten records", Parameters: []ParamInfo{{Name: "record", Doc: "Record to unflatten"}}},
	"upper":          {Label: "upper(value: string) -> string", Doc: "Convert to uppercase", Parameters: []ParamInfo{{Name: "value", Doc: "String to convert"}}},
}

// Aggregate signatures
var aggregateSignatures = map[string]*FunctionSig{
	"and":         {Label: "and(value: bool) -> bool", Doc: "Logical AND of values", Parameters: []ParamInfo{{Name: "value", Doc: "Boolean values"}}},
	"any":         {Label: "any(value: any) -> any", Doc: "Any value from group", Parameters: []ParamInfo{{Name: "value", Doc: "Values to choose from"}}},
	"avg":         {Label: "avg(value: number) -> float64", Doc: "Average of values", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric values"}}},
	"collect":     {Label: "collect(value: any) -> [any]", Doc: "Collect values into array", Parameters: []ParamInfo{{Name: "value", Doc: "Values to collect"}}},
	"collect_map": {Label: "collect_map(key: any, value: any) -> map", Doc: "Collect into map", Parameters: []ParamInfo{{Name: "key", Doc: "Map keys"}, {Name: "value", Doc: "Map values"}}},
	"count":       {Label: "count() -> uint64", Doc: "Count records", Parameters: []ParamInfo{}},
	"dcount":      {Label: "dcount(value: any) -> uint64", Doc: "Distinct count", Parameters: []ParamInfo{{Name: "value", Doc: "Values to count"}}},
	"fuse":        {Label: "fuse(value: any) -> type", Doc: "Fuse schemas in group", Parameters: []ParamInfo{{Name: "value", Doc: "Values to fuse"}}},
	"max":         {Label: "max(value: number) -> number", Doc: "Maximum value", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric values"}}},
	"min":         {Label: "min(value: number) -> number", Doc: "Minimum value", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric values"}}},
	"or":          {Label: "or(value: bool) -> bool", Doc: "Logical OR of values", Parameters: []ParamInfo{{Name: "value", Doc: "Boolean values"}}},
	"sum":         {Label: "sum(value: number) -> number", Doc: "Sum of values", Parameters: []ParamInfo{{Name: "value", Doc: "Numeric values"}}},
	"union":       {Label: "union(value: any) -> set", Doc: "Union of values", Parameters: []ParamInfo{{Name: "value", Doc: "Values to union"}}},
}

// getFunctionSignature returns the signature for a function
func getFunctionSignature(name string) *FunctionSig {
	return functionSignatures[strings.ToLower(name)]
}

// getAggregateSignature returns the signature for an aggregate
func getAggregateSignature(name string) *FunctionSig {
	return aggregateSignatures[strings.ToLower(name)]
}

// getSignatureHelp returns signature help for the current position
func getSignatureHelp(text string, pos Position) *SignatureHelp {
	// Find the function call context
	funcName, paramIndex := findFunctionContext(text, pos)
	if funcName == "" {
		return nil
	}

	funcNameLower := strings.ToLower(funcName)

	// Check functions first
	if sig := functionSignatures[funcNameLower]; sig != nil {
		return buildSignatureHelp(sig, paramIndex)
	}

	// Check aggregates
	if sig := aggregateSignatures[funcNameLower]; sig != nil {
		return buildSignatureHelp(sig, paramIndex)
	}

	return nil
}

// buildSignatureHelp creates a SignatureHelp from a FunctionSig
func buildSignatureHelp(sig *FunctionSig, activeParam int) *SignatureHelp {
	params := make([]ParameterInformation, len(sig.Parameters))

	// Calculate parameter label offsets
	labelOffset := strings.Index(sig.Label, "(") + 1
	currentOffset := labelOffset

	for i, p := range sig.Parameters {
		// Find this parameter in the label
		paramStart := strings.Index(sig.Label[currentOffset:], p.Name)
		if paramStart == -1 {
			continue
		}
		paramStart += currentOffset

		// Find the end of this parameter (comma or closing paren)
		paramEnd := paramStart + len(p.Name)
		for paramEnd < len(sig.Label) && sig.Label[paramEnd] != ',' && sig.Label[paramEnd] != ')' {
			paramEnd++
		}

		params[i] = ParameterInformation{
			Label: [2]int{paramStart, paramEnd},
			Documentation: &MarkupContent{
				Kind:  MarkupKindPlainText,
				Value: p.Doc,
			},
		}

		currentOffset = paramEnd + 1
	}

	if activeParam >= len(params) {
		activeParam = len(params) - 1
	}
	if activeParam < 0 {
		activeParam = 0
	}

	return &SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label: sig.Label,
				Documentation: &MarkupContent{
					Kind:  MarkupKindPlainText,
					Value: sig.Doc,
				},
				Parameters: params,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: activeParam,
	}
}

// findFunctionContext finds the function name and parameter index at position
func findFunctionContext(text string, pos Position) (string, int) {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return "", 0
	}

	// Get text up to cursor position
	var textToCursor strings.Builder
	for i := 0; i <= pos.Line && i < len(lines); i++ {
		if i == pos.Line {
			if pos.Character <= len(lines[i]) {
				textToCursor.WriteString(lines[i][:pos.Character])
			} else {
				textToCursor.WriteString(lines[i])
			}
		} else {
			textToCursor.WriteString(lines[i])
			textToCursor.WriteByte('\n')
		}
	}

	content := textToCursor.String()

	// Walk backward to find matching open paren
	parenDepth := 0
	funcEnd := -1

	for i := len(content) - 1; i >= 0; i-- {
		ch := content[i]
		switch ch {
		case ')':
			parenDepth++
		case '(':
			if parenDepth == 0 {
				funcEnd = i
				break
			}
			parenDepth--
		}
		if funcEnd >= 0 {
			break
		}
	}

	if funcEnd < 0 {
		return "", 0
	}

	// Extract function name
	funcStart := funcEnd - 1
	for funcStart >= 0 && isIdentifierChar(content[funcStart]) {
		funcStart--
	}
	funcStart++

	if funcStart >= funcEnd {
		return "", 0
	}

	funcName := content[funcStart:funcEnd]

	// Count commas to determine parameter index
	paramIndex := 0
	parenDepth = 0
	for i := funcEnd + 1; i < len(content); i++ {
		ch := content[i]
		switch ch {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case ',':
			if parenDepth == 0 {
				paramIndex++
			}
		}
	}

	return funcName, paramIndex
}
