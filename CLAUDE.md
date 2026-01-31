# CLAUDE.md

Project-specific instructions for Claude Code.

## Commit Messages

Follow standard git commit conventions:
- Subject line: max 50 characters, imperative mood
- Body: wrap at 72 characters
- Blank line between subject and body

Example:
```
Fix parser import for new API

Update from ParseQuery to Parse to match the
renamed function in brimdata/super.
```

## Synchronization

Run `/sync` to synchronize with upstream brimdata/super. This updates:
- Keywords, operators, types from PEG grammar
- Functions from function.go
- Aggregates from agg.go
- Go dependency to match synced commit
- Version number based on latest source file date

## Testing

Always run tests before committing LSP changes:
```bash
cd lsp && go build -v && go test -v
```
