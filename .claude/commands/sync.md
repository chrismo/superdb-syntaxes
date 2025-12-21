Perform a full grammar synchronization with upstream brimdata/super.

## Do all of this autonomously:

### 1. Fetch Latest Grammar
Fetch from brimdata/zed main branch:
- `compiler/parser/parser.peg` - keywords, operators, types
- `runtime/sam/expr/function/function.go` - built-in functions
- `runtime/sam/expr/agg/agg.go` - aggregate functions
- Also check `brimdata/zui` `apps/superdb-desktop/src/core/zed-syntax.ts` for reference

### 2. Compare & Update
Compare against local files and update if needed:
- `lsp/completion.go` - add any missing keywords/functions/operators/types
- `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json` - keep TextMate grammar in sync

### 3. Test
Run the full test suite:
```bash
cd lsp && go build -v && go test -v
```
Fix any test failures.

### 4. Build
Build the binary and verify it works:
```bash
cd lsp && go build -o superdb-lsp .
```

### 5. Update Docs
Update `lsp/README.md` with:
- New "Last synchronized" date
- Any new keywords/functions added to the reference section

### 6. Commit & Push
If changes were made:
- Stage all changes
- Commit with a descriptive message listing what was added/changed
- Push to the current branch

### 7. Report
Summarize what was done:
- Number of new items added (by category)
- Test results
- Binary size
- Commit hash
