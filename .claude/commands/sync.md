Perform a full grammar synchronization with upstream brimdata/super.

Use WebFetch or gh instead of curl.

## Do all of this autonomously:

### 1. Fetch Latest Commit & Source Files

**Get the latest commit from main branch** (for Go dependency and versioning):
- `https://api.github.com/repos/brimdata/super/commits?per_page=1`
- Extract the commit SHA and date

**Fetch these files** from brimdata/super main branch to discover new names:

| File | Provides | Why |
|------|----------|-----|
| `compiler/parser/parser.peg` | Keywords, operators, types | The PEG grammar defines the language syntax |
| `runtime/sam/expr/function/function.go` | Built-in scalar functions | Function names are registered at runtime, not in grammar |
| `runtime/sam/expr/agg/agg.go` | Aggregate functions | Aggregate names are registered separately from scalar functions |

### 2. Review Recent Changes

**Get the last sync date** from `lsp/README.md` ("Last synchronized: ...").

**Fetch commits since last sync** to catch signature/behavior changes:
- `https://api.github.com/repos/brimdata/super/commits?since=<last-sync-date>`

Review these commits for changes that affect:
- Function/aggregate signatures (return types, parameters)
- Renamed or removed functions
- New functions not in the registry files

### 3. Compare & Update

Compare against local files and update if needed:
- `lsp/builtins.go` - add any missing keywords/functions/operators/types, update signatures
- `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json` - keep TextMate grammar in sync

### 4. Update Version & Dependencies

Calculate version from the **latest main branch commit date** using format `0.YMMDD`:
- Y = last digit of year (e.g., 2025 → 5)
- MM = 2-digit month
- DD = 2-digit day
- Example: 2025-12-24 → `0.51224`

Update version in:
- `lsp/version.go` - the `Version` constant
- `supersql/spq.tmb/info.plist` - the version string

**Update Go dependency** to the latest main branch commit:
```bash
cd lsp && go get github.com/brimdata/super@<commit-sha> && go mod tidy
```
This ensures the parser used for diagnostics matches the latest upstream version.

### 5. Test

Run the full test suite:
```bash
cd lsp && go build -v && go test -v
```
Fix any test failures.

### 6. Build

Build the binary and verify it works:
```bash
cd lsp && go build -o superdb-lsp .
./superdb-lsp --version
```

### 7. Update Docs

Update `lsp/README.md` with:
- New "Last synchronized" date
- Any new keywords/functions added to the reference section

### 8. Commit & Push

If changes were made:
- Stage all changes
- Commit with a descriptive message listing what was added/changed
- Include the new version number in the commit message
- Push to the current branch

### 9. Report

Summarize what was done:
- New version number
- Number of new items added (by category)
- Any signature changes
- Test results
- Binary size
- Commit hash
