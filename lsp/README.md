# SuperDB LSP Server

Language Server Protocol (LSP) implementation for SuperSQL (SPQ), providing real-time diagnostics and code completion for SuperDB queries.

## Features

- **Diagnostics**: Real-time syntax error detection using the brimdata/super parser
- **Code Completion**: Intelligent suggestions for:
  - Keywords (SQL: `select`, `from`, `where`, `join`, `group`, `order`, etc.)
  - Operators (`sort`, `where`, `yield`, `summarize`, `cut`, `put`, etc.)
  - Functions (`abs`, `ceil`, `floor`, `len`, `split`, `upper`, `cast`, etc.)
  - Aggregate functions (`count`, `sum`, `avg`, `max`, `min`, `collect`, etc.)
  - Types (`int64`, `string`, `bool`, `time`, `duration`, `date`, etc.)
- **Hover**: Documentation on hover for keywords, functions, operators, types, and aggregates
- **Signature Help**: Function parameter hints with documentation as you type
- **Formatting**: Auto-format queries with configurable options (tab size, spaces vs tabs)

## Grammar Synchronization

The LSP server's keyword, function, operator, and type lists are synchronized with [brimdata/super](https://github.com/brimdata/super):

| Source File | Provides | Why |
|-------------|----------|-----|
| `compiler/parser/parser.peg` | Keywords, operators, types | PEG grammar defines language syntax |
| `runtime/sam/expr/function/function.go` | Scalar functions | Function names registered at runtime |
| `runtime/sam/expr/agg/agg.go` | Aggregate functions | Aggregates registered separately |

These files can be updated independently, so version reflects the latest change to any of them.

Last synchronized: December 22, 2025

## Installation

### Prerequisites

- Go 1.22 or later

### Building from Source

```bash
cd lsp
go build -o superdb-lsp .
```

This produces a `superdb-lsp` binary in the current directory.

### Installing Globally

```bash
go install github.com/superdb/superdb-syntaxes/lsp@latest
```

## Usage

The LSP server communicates via stdin/stdout using the Language Server Protocol.

### Quick Test

Try the LSP without editor setup:

```bash
# Build
cd lsp && go build -o superdb-lsp .

# Send an initialize request
echo -e 'Content-Length: 58\r\n\r\n{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
  | ./superdb-lsp 2>/dev/null
```

You should see a JSON response with `capabilities` and `serverInfo`.

### VS Code

Use a generic LSP client extension like [glspc](https://marketplace.visualstudio.com/items?itemName=APerezMuñoz.lsp-client).

Add to your `settings.json`:

```json
{
  "glspc.serverCommand": "/path/to/superdb-lsp",
  "glspc.languageId": "spq"
}
```

### Neovim (with nvim-lspconfig)

```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.superdb then
  configs.superdb = {
    default_config = {
      cmd = { '/path/to/superdb-lsp' },
      filetypes = { 'spq', 'supersql' },
      root_dir = function(fname)
        return lspconfig.util.find_git_ancestor(fname) or vim.fn.getcwd()
      end,
      settings = {},
    },
  }
end

lspconfig.superdb.setup{}
```

### Sublime Text (with LSP package)

Add to `LSP.sublime-settings`:

```json
{
  "clients": {
    "superdb": {
      "enabled": true,
      "command": ["/path/to/superdb-lsp"],
      "selector": "source.spq"
    }
  }
}
```

### Emacs (with lsp-mode)

```elisp
(require 'lsp-mode)

(add-to-list 'lsp-language-id-configuration '(spq-mode . "spq"))

(lsp-register-client
 (make-lsp-client
  :new-connection (lsp-stdio-connection '("/path/to/superdb-lsp"))
  :major-modes '(spq-mode)
  :server-id 'superdb-lsp))
```

### Helix

Add to `~/.config/helix/languages.toml`:

```toml
[[language]]
name = "spq"
scope = "source.spq"
file-types = ["spq"]
language-servers = ["superdb-lsp"]

[language-server.superdb-lsp]
command = "/path/to/superdb-lsp"
```

## LSP Capabilities

### Supported Methods

| Method | Description |
|--------|-------------|
| `initialize` | Handshake with capabilities negotiation |
| `initialized` | Confirmation of initialization |
| `shutdown` | Graceful shutdown request |
| `exit` | Server termination |
| `textDocument/didOpen` | Document opened notification |
| `textDocument/didChange` | Document changed notification |
| `textDocument/didClose` | Document closed notification |
| `textDocument/completion` | Code completion request |
| `textDocument/hover` | Hover documentation request |
| `textDocument/signatureHelp` | Function signature help request |
| `textDocument/formatting` | Document formatting request |

### Server Capabilities

- **Text Document Sync**: Full document sync (mode 1)
- **Completion Provider**: Triggered by `.`, `|`, `(`, `:`, `=`
- **Hover Provider**: Documentation for keywords, functions, types, operators
- **Signature Help Provider**: Triggered by `(` and `,`
- **Document Formatting Provider**: Formats queries with configurable options

## Development

### Running Tests

```bash
cd lsp
go test -v
```

### Debug Mode

The server logs to stderr, so you can capture logs:

```bash
./superdb-lsp 2> lsp.log
```

## Architecture

```
lsp/
├── main.go          # Entry point and server loop
├── protocol.go      # LSP protocol types
├── handlers.go      # Request/notification handlers
├── diagnostics.go   # Parsing and diagnostic generation
├── completion.go    # Completion item generation
├── hover.go         # Hover documentation
├── signature.go     # Function signature help
├── format.go        # Document formatting
├── server_test.go   # Test harness
└── go.mod           # Go module definition
```

## SuperSQL Reference

### Core Keywords
`const`, `file`, `from`, `func`, `let`, `op`, `this`, `type`

### SQL Keywords
`select`, `as`, `by`, `where`, `group`, `having`, `order`, `limit`, `offset`, `with`, `distinct`, `all`, `aggregate`

### Join Keywords
`join`, `inner`, `left`, `right`, `outer`, `full`, `cross`, `anti`, `on`, `using`

### Logic Keywords
`and`, `or`, `not`, `in`, `like`, `is`, `between`

### Control Flow
`case`, `when`, `then`, `else`, `end`, `default`

### Literals
`true`, `false`, `null`

### Other Keywords
`asc`, `at`, `call`, `cast`, `desc`, `enum`, `error`, `exists`, `extract`, `fn`, `for`, `lambda`, `materialized`, `nulls`, `first`, `last`, `ordinality`, `pragma`, `recursive`, `shape`, `shapes`, `substring`, `union`, `value`

### Operators
`assert`, `cut`, `debug`, `drop`, `explode`, `fork`, `fuse`, `head`, `join`, `load`, `merge`, `output`, `over`, `pass`, `put`, `rename`, `sample`, `search`, `skip`, `sort`, `summarize`, `switch`, `tail`, `top`, `uniq`, `unnest`, `values`, `where`, `yield`

### Functions
`abs`, `base64`, `bucket`, `cast`, `ceil`, `cidr_match`, `coalesce`, `compare`, `date_part`, `error`, `fields`, `flatten`, `floor`, `grep`, `grok`, `has`, `has_error`, `hex`, `is`, `is_error`, `join`, `kind`, `ksuid`, `len`, `length`, `levenshtein`, `log`, `lower`, `max`, `min`, `missing`, `nameof`, `nest_dotted`, `network_of`, `now`, `nullif`, `parse_sup`, `parse_uri`, `position`, `pow`, `quiet`, `regexp`, `regexp_replace`, `replace`, `round`, `split`, `sqrt`, `strftime`, `trim`, `typename`, `typeof`, `under`, `unflatten`, `upper`

### Aggregates
`and`, `any`, `avg`, `collect`, `collect_map`, `count`, `dcount`, `fuse`, `max`, `min`, `or`, `sum`, `union`

### Types
`uint8`, `uint16`, `uint32`, `uint64`, `uint128`, `uint256`, `int8`, `int16`, `int32`, `int64`, `int128`, `int256`, `float16`, `float32`, `float64`, `float128`, `float256`, `decimal32`, `decimal64`, `decimal128`, `decimal256`, `duration`, `time`, `date`, `timestamp`, `bool`, `bytes`, `string`, `ip`, `net`, `type`, `null`

### SQL Type Aliases
`bigint`, `smallint`, `boolean`, `text`, `bytea`

## License

See the repository root for license information.
