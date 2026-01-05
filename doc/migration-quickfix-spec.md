# Migration Quick-Fix Spec

LSP Code Actions to help users migrate from deprecated zq syntax to current SuperDB syntax.

Reference: https://github.com/chrismo/superkit/blob/main/doc/zq-to-super-upgrades.md

## Overview

When the LSP detects deprecated syntax, it will:
1. Report a diagnostic (warning or error)
2. Attach a Code Action with an automatic fix
3. Support "Fix all in file" for batch migrations

## Keyword Renames

Simple token replacements.

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `deprecated-yield` | `yield` | `values` | `'yield' is deprecated, use 'values'` |
| `deprecated-func` | `func` | `fn` | `'func' is deprecated, use 'fn'` |
| `deprecated-over` | `over` | `unnest` | `'over' is deprecated, use 'unnest'` |
| `deprecated-arrow` | `=>` | `into` | `'=>' is deprecated, use 'into'` |

### Example

```
Input:  yield x
Fix:    values x
```

## Comment Syntax

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `deprecated-comment-slash` | `// comment` | `-- comment` | `'//' comments are deprecated, use '--'` |

### Example

```
Input:  // this is a comment
Fix:    -- this is a comment
```

## Function Renames

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `deprecated-parse-zson` | `parse_zson` | `parse_sup` | `'parse_zson' is deprecated, use 'parse_sup'` |

## Implicit `this` Argument

Functions no longer imply `this` as the first argument.

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `implicit-this-grep` | `grep(/pattern/)` | `grep('pattern', this)` | `grep() requires explicit 'this' argument` |
| `implicit-this-is` | `is(<type>)` | `is(this, <type>)` | `is() requires explicit 'this' argument` |
| `implicit-this-nest-dotted` | `nest_dotted()` | `nest_dotted(this)` | `nest_dotted() requires explicit 'this' argument` |

### Detection

These require semantic analysis—check function call arity against known signatures.

### Example

```
Input:  grep(/error/)
Fix:    grep('error', this)

Input:  is(<string>)
Fix:    is(this, <string>)

Input:  nest_dotted()
Fix:    nest_dotted(this)
```

## Casting Syntax

Direct function-style casting replaced with PostgreSQL-style cast operator.

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `deprecated-cast-time` | `time('...')` | `'...'::time` | `Function-style cast deprecated, use '::time'` |
| `deprecated-cast-duration` | `duration('...')` | `'...'::duration` | `Function-style cast deprecated, use '::duration'` |
| `deprecated-cast-ip` | `ip('...')` | `'...'::ip` | `Function-style cast deprecated, use '::ip'` |
| `deprecated-cast-net` | `net('...')` | `'...'::net` | `Function-style cast deprecated, use '::net'` |

### Detection

Detect calls to type-name functions with string literal arguments.

### Example

```
Input:  time('2025-08-28T00:00:00Z')
Fix:    '2025-08-28T00:00:00Z'::time
```

## User-Defined Operators

Operator declaration syntax changed.

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `deprecated-op-parens` | `op name(args):` | `op name args:` | `Operator declaration no longer uses parentheses` |

### Example

```
Input:  op components(s): (
          ...
        )
        components(this)

Fix:    op components s: (
          ...
        )
        components this
```

Note: This also affects operator invocations—parentheses removed from calls.

## Removed Functions

These have no direct replacement—show error with guidance.

| Diagnostic Code | Function | Message |
|-----------------|----------|---------|
| `removed-crop` | `crop()` | `'crop()' was removed, use explicit casting` |
| `removed-fill` | `fill()` | `'fill()' was removed, use explicit casting` |
| `removed-fit` | `fit()` | `'fit()' was removed, use explicit casting` |
| `removed-order` | `order()` | `'order()' was removed, use explicit casting` |
| `removed-shape` | `shape()` | `'shape()' was removed, use explicit casting` |

No automatic fix available—these require manual refactoring.

## Streaming Aggregation (PR 6355)

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `deprecated-put-agg` | `put row_num:=count(this)` | `count {row,...this}` | `Streaming aggregation syntax changed` |

### Detection

Detect `put field:=agg(...)` pattern.

### Complexity

Medium—requires understanding the aggregation context. May not be fully automatable.

## SQL FROM Syntax (PR 6405)

| Diagnostic Code | Old | New | Message |
|-----------------|-----|-----|---------|
| `pipe-from-select` | `from (values ...)` | `select * from (values ...)` | `Pipe 'from' requires explicit 'select *'` |

### Detection

Detect `from` at start of pipe expression with subquery.

## Implementation Priority

### Phase 1: Simple Token Replacements
- [ ] `yield` → `values`
- [ ] `func` → `fn`
- [ ] `over` → `unnest`
- [ ] `=>` → `into`
- [ ] `//` → `--` comments
- [ ] `parse_zson` → `parse_sup`

### Phase 2: Function Signature Changes
- [ ] Implicit `this` detection for `grep`, `is`, `nest_dotted`
- [ ] Cast syntax migration (`time()` → `::time`, etc.)

### Phase 3: Structural Changes
- [ ] Operator declaration/invocation syntax
- [ ] Streaming aggregation patterns
- [ ] SQL `from` syntax

### Phase 4: Removal Warnings
- [ ] Removed functions (`crop`, `fill`, `fit`, `order`, `shape`)

## Code Action Kinds

```
quickfix              - Individual fix for one diagnostic
source.fixAll         - Fix all auto-fixable issues in file
source.fixAll.migrate - Fix all migration issues specifically
```

## Configuration (Future)

Consider adding LSP initialization options:

```json
{
  "superdb.migration.enabled": true,
  "superdb.migration.severity": "warning",
  "superdb.migration.targetVersion": "0.51231"
}
```

This would allow users to:
- Disable migration warnings if working with intentionally old syntax
- Choose error vs warning severity
- Target a specific version for migration suggestions
