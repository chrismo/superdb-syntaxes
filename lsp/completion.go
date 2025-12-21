package main

import (
	"strings"
)

// SuperSQL keywords - from PEG grammar
var keywords = []struct {
	name   string
	detail string
}{
	// Core keywords
	{"const", "Declare a constant"},
	{"file", "File source"},
	{"from", "Data source"},
	{"func", "Define a function"},
	{"op", "Define an operator"},
	{"this", "Current value reference"},
	{"type", "Type definition"},
	{"let", "Variable binding"},
	// SQL keywords
	{"select", "Select fields"},
	{"as", "Alias"},
	{"by", "Group by field"},
	{"where", "Filter condition"},
	{"group", "Group records"},
	{"having", "Filter groups"},
	{"order", "Order results"},
	{"limit", "Limit results"},
	{"offset", "Skip results"},
	{"with", "Common table expression"},
	{"distinct", "Distinct values"},
	{"all", "All values"},
	// Join keywords
	{"join", "Join data sources"},
	{"inner", "Inner join"},
	{"left", "Left join"},
	{"right", "Right join"},
	{"outer", "Outer join"},
	{"full", "Full join"},
	{"cross", "Cross join"},
	{"anti", "Anti join"},
	{"on", "Join condition"},
	{"using", "Join using columns"},
	// Logic keywords
	{"and", "Logical AND"},
	{"or", "Logical OR"},
	{"not", "Logical NOT"},
	{"in", "In set"},
	{"like", "Pattern match"},
	{"is", "Type check"},
	{"between", "Range check"},
	// Control flow
	{"case", "Case expression"},
	{"when", "Case condition"},
	{"then", "Case result"},
	{"else", "Default case"},
	{"end", "End case"},
	{"default", "Default branch"},
	// Literals
	{"true", "Boolean true"},
	{"false", "Boolean false"},
	{"null", "Null value"},
	// Other keywords
	{"aggregate", "Aggregate expression"},
	{"nulls", "Null ordering"},
	{"first", "First value"},
	{"last", "Last value"},
	{"asc", "Sort ascending"},
	{"desc", "Sort descending"},
	{"at", "At location/time"},
	{"call", "Function call"},
	{"cast", "Type cast"},
	{"enum", "Enumeration type"},
	{"error", "Error value"},
	{"exists", "SQL EXISTS"},
	{"extract", "Extract component"},
	{"fn", "Function shorthand"},
	{"for", "For iteration"},
	{"lambda", "Lambda expression"},
	{"materialized", "Materialized view"},
	{"ordinality", "WITH ORDINALITY"},
	{"pragma", "Compiler directive"},
	{"recursive", "Recursive CTE"},
	{"shape", "Value shape"},
	{"shapes", "Get shapes"},
	{"substring", "Substring function"},
	{"union", "SQL UNION"},
	{"value", "Value keyword"},
}

// Built-in operators/ops - from PEG grammar and zui
var operators = []struct {
	name   string
	detail string
}{
	{"assert", "Assert condition"},
	{"combine", "Combine multiple streams"},
	{"cut", "Select and reorder fields"},
	{"debug", "Debug output"},
	{"drop", "Remove fields from records"},
	{"explode", "Explode array into records"},
	{"file", "Read from file"},
	{"fork", "Fork the data flow"},
	{"from", "Specify data source"},
	{"fuse", "Fuse schemas together"},
	{"get", "HTTP GET request"},
	{"head", "Take first N records"},
	{"join", "Join two data sources"},
	{"load", "Load data into pool"},
	{"merge", "Merge sorted streams"},
	{"output", "Output to destination"},
	{"over", "Iterate over values"},
	{"pass", "Pass through unchanged"},
	{"put", "Add/update fields"},
	{"rename", "Rename fields"},
	{"sample", "Sample random records"},
	{"search", "Search expression"},
	{"skip", "Skip N records"},
	{"sort", "Sort records"},
	{"summarize", "Aggregate data"},
	{"switch", "Conditional branching"},
	{"tail", "Take last N records"},
	{"top", "Top N by field"},
	{"uniq", "Remove duplicates"},
	{"unnest", "Unnest nested values"},
	{"values", "Extract values"},
	{"where", "Filter records"},
	{"yield", "Output values"},
}

// Built-in functions - from brimdata/zed function.go
var functions = []struct {
	name   string
	detail string
}{
	{"abs", "Absolute value"},
	{"base64", "Base64 encode/decode"},
	{"bucket", "Bucket values into ranges"},
	{"cast", "Cast value to type"},
	{"ceil", "Ceiling function"},
	{"cidr_match", "Match IP against CIDR"},
	{"coalesce", "First non-null value"},
	{"compare", "Compare two values"},
	{"crop", "Crop value to type"},
	{"date_part", "Extract date component"},
	{"error", "Create error value"},
	{"every", "Time bucket interval"},
	{"fields", "Get record field names"},
	{"fill", "Fill null values"},
	{"flatten", "Flatten nested records"},
	{"floor", "Floor function"},
	{"grep", "Search with pattern"},
	{"grok", "Parse with grok pattern"},
	{"has", "Check if field exists"},
	{"has_error", "Check for error"},
	{"hex", "Hexadecimal conversion"},
	{"is", "Type check"},
	{"is_error", "Check if value is error"},
	{"join", "Join strings"},
	{"kind", "Get value kind"},
	{"ksuid", "Generate KSUID"},
	{"len", "Length of value"},
	{"length", "Length of value (alias)"},
	{"levenshtein", "Levenshtein distance"},
	{"log", "Logarithm"},
	{"lower", "Convert to lowercase"},
	{"map", "Map function over array"},
	{"max", "Maximum of values"},
	{"min", "Minimum of values"},
	{"missing", "Create missing value"},
	{"nameof", "Get type name"},
	{"nest_dotted", "Nest dotted field names"},
	{"network_of", "Get network from IP"},
	{"now", "Current timestamp"},
	{"nullif", "Return null if equal"},
	{"order", "Order type info"},
	{"parse_sup", "Parse Super format"},
	{"parse_uri", "Parse URI string"},
	{"parse_zson", "Parse ZSON string"},
	{"position", "Find substring position"},
	{"pow", "Power function"},
	{"quiet", "Suppress errors"},
	{"regexp", "Regular expression match"},
	{"regexp_replace", "Regex replacement"},
	{"replace", "String replacement"},
	{"round", "Round to precision"},
	{"rune_len", "UTF-8 rune length"},
	{"shape", "Get value shape"},
	{"split", "Split string"},
	{"sqrt", "Square root"},
	{"strftime", "Format time as string"},
	{"trim", "Trim whitespace"},
	{"typename", "Get type name"},
	{"typeof", "Get type of value"},
	{"typeunder", "Get underlying type"},
	{"under", "Get underlying value"},
	{"unflatten", "Unflatten records"},
	{"upper", "Convert to uppercase"},
}

// Built-in aggregate functions - from brimdata/zed agg.go
var aggregates = []struct {
	name   string
	detail string
}{
	{"and", "Logical AND of values"},
	{"any", "Any value from group"},
	{"avg", "Average of values"},
	{"collect", "Collect values into array"},
	{"collect_map", "Collect into map"},
	{"count", "Count records"},
	{"dcount", "Distinct count"},
	{"fuse", "Fuse schemas in group"},
	{"max", "Maximum value"},
	{"min", "Minimum value"},
	{"or", "Logical OR of values"},
	{"sum", "Sum of values"},
	{"union", "Union of values"},
}

// Built-in types - from PEG grammar
var types = []struct {
	name   string
	detail string
}{
	// Unsigned integers
	{"uint8", "8-bit unsigned integer"},
	{"uint16", "16-bit unsigned integer"},
	{"uint32", "32-bit unsigned integer"},
	{"uint64", "64-bit unsigned integer"},
	{"uint128", "128-bit unsigned integer"},
	{"uint256", "256-bit unsigned integer"},
	// Signed integers
	{"int8", "8-bit signed integer"},
	{"int16", "16-bit signed integer"},
	{"int32", "32-bit signed integer"},
	{"int64", "64-bit signed integer"},
	{"int128", "128-bit signed integer"},
	{"int256", "256-bit signed integer"},
	// Floats
	{"float16", "16-bit float"},
	{"float32", "32-bit float"},
	{"float64", "64-bit float"},
	{"float128", "128-bit float"},
	{"float256", "256-bit float"},
	// Decimals
	{"decimal32", "32-bit decimal"},
	{"decimal64", "64-bit decimal"},
	{"decimal128", "128-bit decimal"},
	{"decimal256", "256-bit decimal"},
	// Time types
	{"duration", "Duration type"},
	{"time", "Timestamp type"},
	{"date", "Date type"},
	{"timestamp", "Timestamp type (alias)"},
	// Other types
	{"bool", "Boolean type"},
	{"bytes", "Byte array type"},
	{"string", "String type"},
	{"ip", "IP address type"},
	{"net", "Network CIDR type"},
	{"type", "Type type"},
	{"null", "Null type"},
	// SQL type aliases
	{"bigint", "64-bit integer (alias for int64)"},
	{"smallint", "16-bit integer (alias for int16)"},
	{"boolean", "Boolean (alias for bool)"},
	{"text", "Text (alias for string)"},
	{"bytea", "Byte array (alias for bytes)"},
}

// getCompletions returns completion items based on the current context
func getCompletions(text string, pos Position) []CompletionItem {
	var items []CompletionItem

	// Get the current line and word being typed
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return items
	}

	line := lines[pos.Line]
	prefix := ""
	if pos.Character <= len(line) {
		// Get the word prefix before cursor
		start := pos.Character
		for start > 0 && isIdentifierChar(line[start-1]) {
			start--
		}
		if start < pos.Character {
			prefix = strings.ToLower(line[start:pos.Character])
		}
	}

	// Check context for better completions
	context := getCompletionContext(line, pos.Character)

	// Add completions based on context
	switch context {
	case contextType:
		// After type-related keywords, suggest types
		items = append(items, getTypeCompletions(prefix)...)
	case contextFunction:
		// After opening paren or in function context
		items = append(items, getFunctionCompletions(prefix)...)
		items = append(items, getAggregateCompletions(prefix)...)
	default:
		// General context - suggest everything
		items = append(items, getKeywordCompletions(prefix)...)
		items = append(items, getOperatorCompletions(prefix)...)
		items = append(items, getFunctionCompletions(prefix)...)
		items = append(items, getAggregateCompletions(prefix)...)
		items = append(items, getTypeCompletions(prefix)...)
	}

	return items
}

type completionContext int

const (
	contextGeneral completionContext = iota
	contextType
	contextFunction
)

// getCompletionContext analyzes the line to determine the completion context
func getCompletionContext(line string, col int) completionContext {
	if col > len(line) {
		col = len(line)
	}
	prefix := strings.ToLower(line[:col])

	// Check if we're after a type cast operator
	if strings.Contains(prefix, "cast(") ||
		strings.Contains(prefix, "::") ||
		strings.HasSuffix(strings.TrimSpace(prefix), "<") {
		return contextType
	}

	// Check if we're inside a function call
	openParens := strings.Count(prefix, "(") - strings.Count(prefix, ")")
	if openParens > 0 {
		return contextFunction
	}

	return contextGeneral
}

func isIdentifierChar(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '_'
}

func getKeywordCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, kw := range keywords {
		if prefix == "" || strings.HasPrefix(strings.ToLower(kw.name), prefix) {
			items = append(items, CompletionItem{
				Label:  kw.name,
				Kind:   CompletionItemKindKeyword,
				Detail: kw.detail,
			})
		}
	}
	return items
}

func getOperatorCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, op := range operators {
		if prefix == "" || strings.HasPrefix(strings.ToLower(op.name), prefix) {
			items = append(items, CompletionItem{
				Label:  op.name,
				Kind:   CompletionItemKindFunction,
				Detail: "operator: " + op.detail,
			})
		}
	}
	return items
}

func getFunctionCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, fn := range functions {
		if prefix == "" || strings.HasPrefix(strings.ToLower(fn.name), prefix) {
			items = append(items, CompletionItem{
				Label:      fn.name,
				Kind:       CompletionItemKindFunction,
				Detail:     "function: " + fn.detail,
				InsertText: fn.name + "($1)",
			})
		}
	}
	return items
}

func getAggregateCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, agg := range aggregates {
		if prefix == "" || strings.HasPrefix(strings.ToLower(agg.name), prefix) {
			items = append(items, CompletionItem{
				Label:      agg.name,
				Kind:       CompletionItemKindFunction,
				Detail:     "aggregate: " + agg.detail,
				InsertText: agg.name + "($1)",
			})
		}
	}
	return items
}

func getTypeCompletions(prefix string) []CompletionItem {
	var items []CompletionItem
	for _, t := range types {
		if prefix == "" || strings.HasPrefix(strings.ToLower(t.name), prefix) {
			items = append(items, CompletionItem{
				Label:  t.name,
				Kind:   CompletionItemKindClass,
				Detail: "type: " + t.detail,
			})
		}
	}
	return items
}
