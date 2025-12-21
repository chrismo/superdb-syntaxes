# SuperDB Syntax Highlighting

Syntax highlighting support for [SuperDB](https://github.com/brimdata/super) query languages:

- **SPQ (SuperSQL)** - The SuperDB query language
- **ZSON** - SuperDB's JSON-like data format

## Supported Editors

| Editor | Method | Status |
|--------|--------|--------|
| JetBrains IDEs | TextMate Bundles plugin | Working |
| VS Code | Extension (coming soon) | Planned |
| Sublime Text | TextMate bundle | Should work |
| GitHub | Linguist PR | Planned |

## Installation

### JetBrains IDEs (IntelliJ, RubyMine, WebStorm, etc.)

1. Install the [TextMate Bundles](https://plugins.jetbrains.com/plugin/7221-textmate-bundles) plugin
2. Go to **Settings** → **Editor** → **TextMate Bundles**
3. Click **+** and select the `supersql/spq.tmb` folder from this repository
4. Optionally add `zson/zson.tmb` for ZSON support
5. Click **Apply** and **OK**

**Note:** After making changes to the grammar, you must reload the bundle:
```
Settings → TextMate Bundles → Remove → Apply → Re-add → Close
```

### Sublime Text

Copy the `.tmb` bundle folders to your Sublime Text Packages directory.

## SPQ Language Features

The grammar supports the full SuperSQL language as defined in the
[Super compiler](https://github.com/brimdata/super/tree/main/compiler/parser):

### Comments
```sql
// Single-line comment
-- SQL-style comment
/* Multi-line
   block comment */
```

### Declarations
```sql
const pi = 3.14159
func double(x): (x * 2)
op myOperator(): (over this | yield {result: x})
```

### Pipeline Operators
```sql
from "data.json"
  | where status == "active"
  | cut name, email
  | sort -timestamp
  | head 10
```

### SQL Syntax
```sql
SELECT id, name, COUNT(*) AS total
FROM users
WHERE status = 'active'
GROUP BY category
HAVING total > 5
ORDER BY name ASC
LIMIT 100
```

### F-Strings (Template Strings)
```sql
f"Hello, {name}!"
f'Count: {count(this)}'
```

### Types
```sql
uint8, uint16, uint32, uint64, uint128, uint256
int8, int16, int32, int64, int128, int256
float16, float32, float64, float128, float256
decimal32, decimal64, decimal128, decimal256
bool, bytes, string, ip, net, type, null, duration, time, error
```

### Operators
```sql
|>          // Pipe
:=          // Assignment
:: :>       // Type cast
...         // Spread
== != <>    // Comparison
&& || !     // Logical
```

## Grammar Structure

The TextMate grammar (`spq.tmLanguage.json`) is organized into these scopes:

| Scope | Description |
|-------|-------------|
| `comment.*` | Comments (`//`, `--`, `/* */`) |
| `string.*` | String literals and f-strings |
| `constant.language.*` | `true`, `false`, `null` |
| `constant.numeric.*` | Numbers (int, float, hex) |
| `keyword.control.*` | `const`, `func`, `op`, `type` |
| `keyword.other.sql.*` | SQL keywords |
| `keyword.operator.*` | Operators |
| `support.function.*` | Built-in functions |
| `support.function.operator.*` | Pipeline operators |
| `support.function.aggregate.*` | Aggregate functions |
| `support.type.*` | Type names |

## Contributing

### Testing Changes

1. Edit `supersql/spq.tmb/Syntaxes/spq.tmLanguage.json`
2. Use `supersql/sample.spq` to verify highlighting
3. Reload in your editor (see installation notes above)

### Adding New Keywords

Keywords are defined in the `repository` section of the grammar file. To add a new keyword:

1. Find the appropriate category (`functions-builtin`, `sql-keywords`, etc.)
2. Add the keyword to the regex pattern
3. Test with a sample file

### Reference

- [TextMate Grammar Docs](https://macromates.com/manual/en/language_grammars)
- [Super Parser PEG](https://github.com/brimdata/super/blob/main/compiler/parser/parser.peg)
- [TextMate Naming Conventions](https://macromates.com/manual/en/language_grammars#naming_conventions)

## Roadmap

- [ ] VS Code extension
- [ ] GitHub Linguist PR for `.spq` file highlighting
- [ ] Language Server Protocol (LSP) for rich IDE features
- [ ] IntelliJ native plugin

## Related Projects

- [SuperDB](https://github.com/brimdata/super) - The database engine
- [Zui](https://github.com/brimdata/zui) - Desktop app for SuperDB

## License

MIT
