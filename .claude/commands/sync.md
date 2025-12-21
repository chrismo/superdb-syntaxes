# Sync Grammar with SuperDB PEG Spec

Perform a comprehensive sync of the SPQ TextMate grammar with the latest SuperDB PEG parser specification.

## Steps to Execute

### 1. Fetch Latest PEG Spec
Fetch the current PEG grammar from: https://raw.githubusercontent.com/brimdata/super/main/compiler/parser/parser.peg

Extract and document:
- All keywords (CONST, FN, OP, LET, PRAGMA, TYPE, LAMBDA, etc.)
- All pipeline operators (from, where, sort, head, etc.)
- All aggregate functions (count, sum, avg, min, max, first, last, etc.)
- All type names (uint8-64, int8-64, float16-64, bool, string, etc.)
- All SQL keywords
- Comment syntax
- String syntax (regular, raw, f-strings, backtick)
- Number formats
- Duration formats
- All operators

### 2. Compare with Current Grammar
Read the current grammar at: `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json`

Identify:
- Keywords/operators in PEG but missing from grammar (need to ADD)
- Keywords/operators in grammar but removed from PEG (need to REMOVE)
- Any syntax changes

### 3. Update Grammar
Make necessary updates to `spq.tmLanguage.json`:
- Add new keywords/operators/functions
- Remove deprecated ones
- Fix any patterns that don't match the PEG

### 4. Update Tests
Ensure `test/syntax.test.spq` covers:
- All keywords and operators
- All types
- All string formats
- All number formats
- All duration formats

Add tests for any new syntax, remove tests for removed syntax.

### 5. Run Validation
```bash
npm run ci
```

This runs:
- JSON validation
- Grammar structure validation
- All syntax tests (vscode-tmgrammar-test)

### 6. Update Sample File
Update `supersql/sample.spq` to demonstrate all current syntax features.

### 7. Report Results
Provide a summary of:
- What was added
- What was removed
- What was changed
- Test results (pass/fail count)
- Any issues found

### 8. Commit (if all tests pass)
If all tests pass and changes were made, commit with a message like:
```
Sync grammar with PEG spec (YYYY-MM-DD)

Added: [list new items]
Removed: [list removed items]
Changed: [list changes]

All 211+ tests passing.
```

## Important Notes
- SuperDB is in pre-release, expect breaking changes
- The PEG is the source of truth
- Case sensitivity matters: lowercase = SPQ operators, UPPERCASE = SQL keywords
- Always run tests before committing
