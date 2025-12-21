Perform a full grammar synchronization with upstream brimdata/super.

## Do all of this autonomously:

### 1. Fetch Latest Grammar & Version
Fetch from brimdata/zed main branch:
- `compiler/parser/parser.peg` - keywords, operators, types
- `runtime/sam/expr/function/function.go` - built-in functions
- `runtime/sam/expr/agg/agg.go` - aggregate functions
- Also check `brimdata/zui` `apps/superdb-desktop/src/core/zed-syntax.ts` for reference
- **Get the latest commit date** from `https://api.github.com/repos/brimdata/zed/commits/main`

### 2. Compare & Update
Compare against local files and update if needed:
- `lsp/completion.go` - add any missing keywords/functions/operators/types
- `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json` - keep TextMate grammar in sync

### 3. Update Version
Calculate version from upstream's latest commit date using format `0.YMMDD`:
- Y = last digit of year (e.g., 2025 → 5)
- MM = 2-digit month
- DD = 2-digit day
- Example: 2025-12-18 → `0.51218`

Update version in:
- `lsp/main.go` - the `Version` constant
- `supersql/spq.tmb/info.plist` - the version string

### 4. Test
Run the full test suite:
```bash
cd lsp && go build -v && go test -v
```
Fix any test failures.

### 5. Build
Build the binary and verify it works:
```bash
cd lsp && go build -o superdb-lsp .
./superdb-lsp --version
```

### 6. Update Docs
Update `lsp/README.md` with:
- New "Last synchronized" date
- Any new keywords/functions added to the reference section

### 7. Commit & Push
If changes were made:
- Stage all changes
- Commit with a descriptive message listing what was added/changed
- Include the new version number in the commit message
- Push to the current branch

### 8. Report
Summarize what was done:
- New version number
- Number of new items added (by category)
- Test results
- Binary size
- Commit hash
