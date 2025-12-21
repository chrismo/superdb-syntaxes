# Contributing to SuperDB Syntax Highlighting

Thank you for your interest in contributing! This document outlines how to work with the syntax grammars.

## Project Structure

```
superdb-syntaxes/
├── supersql/
│   ├── spq.tmb/                    # TextMate bundle
│   │   ├── info.plist              # Bundle metadata
│   │   └── Syntaxes/
│   │       └── spq.tmLanguage.json # Main grammar file
│   └── sample.spq                  # Test/sample file
├── zson/
│   ├── zson.tmb/                   # ZSON TextMate bundle
│   │   ├── info.plist
│   │   └── Syntaxes/
│   │       └── zson.tmLanguage     # ZSON grammar (XML plist)
│   ├── sample.json
│   └── sample.zson
├── README.md
├── CONTRIBUTING.md
└── NOTES.md                        # Development notes
```

## Grammar Format

The grammars use TextMate format, which is supported by many editors. Two formats exist:

1. **JSON** (`.tmLanguage.json`) - Easier to edit, used for SPQ
2. **XML plist** (`.tmLanguage`) - Traditional format, used for ZSON

## Adding New Syntax Features

### 1. Check the Parser PEG

First, verify the syntax in the [Super parser](https://github.com/brimdata/super/blob/main/compiler/parser/parser.peg).

### 2. Edit the Grammar

Open `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json` and find the appropriate section:

```json
{
  "repository": {
    "keywords": { ... },           // const, func, op, type
    "sql-keywords": { ... },       // SELECT, FROM, WHERE, etc.
    "operators-builtin": { ... },  // Pipeline operators
    "functions-builtin": { ... },  // Built-in functions
    "functions-aggregate": { ... }, // Aggregate functions
    "types": { ... },              // Type names
    "operators": { ... },          // Symbolic operators
    "numbers": { ... },            // Numeric literals
    "strings": { ... },            // String literals
    "fstrings": { ... },           // F-string templates
    "comments": { ... },           // Comments
    "constants": { ... }           // true, false, null
  }
}
```

### 3. Add Your Pattern

Example - adding a new built-in function `my_func`:

```json
"functions-builtin": {
  "name": "support.function.spq",
  "match": "\\b(abs|base64|...|my_func|...)\\b"
}
```

### 4. Test Your Changes

1. Add test cases to `supersql/sample.spq`
2. Reload in your editor
3. Verify highlighting is correct

## Scope Naming Conventions

Use standard TextMate scopes for editor compatibility:

| Use Case | Scope Pattern |
|----------|---------------|
| Keywords | `keyword.control.*`, `keyword.other.*` |
| Operators | `keyword.operator.*` |
| Functions | `support.function.*` |
| Types | `support.type.*` |
| Strings | `string.quoted.*`, `string.interpolated.*` |
| Numbers | `constant.numeric.*` |
| Comments | `comment.line.*`, `comment.block.*` |
| Booleans/null | `constant.language.*` |

See [TextMate Naming Conventions](https://macromates.com/manual/en/language_grammars#naming_conventions).

## Testing in Editors

### JetBrains IDEs

Reload the bundle after changes:
```
Settings → Editor → TextMate Bundles → Remove → Apply → Re-add → Close
```

### VS Code (future)

```bash
# From the extension directory
npm run test
```

### Sublime Text

Close and reopen the editor, or use Package Control to reload.

## Common Patterns

### Word Boundaries

Always use `\b` for word boundaries:
```json
"match": "\\b(keyword1|keyword2)\\b"
```

### Case Insensitivity (SQL keywords)

Use `(?i)` flag:
```json
"match": "(?i)\\b(SELECT|FROM|WHERE)\\b"
```

### Multi-line Constructs

Use `begin`/`end` patterns:
```json
{
  "begin": "/\\*",
  "end": "\\*/",
  "name": "comment.block.spq"
}
```

### Nested Patterns

Use `patterns` array inside begin/end:
```json
{
  "begin": "\"",
  "end": "\"",
  "patterns": [
    { "include": "#string-escapes" }
  ]
}
```

## Pull Request Checklist

- [ ] Changes match the Super parser PEG
- [ ] Test cases added to `sample.spq`
- [ ] Tested in at least one editor
- [ ] Scopes follow naming conventions
- [ ] No breaking changes to existing highlighting
