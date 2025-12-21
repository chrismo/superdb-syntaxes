# SuperDB LSP Server

Language Server Protocol (LSP) implementation for SuperSQL (SPQ), providing real-time diagnostics and code completion for SuperDB queries.

## Features

- **Diagnostics**: Real-time syntax error detection using the brimdata/zed parser
- **Code Completion**: Intelligent suggestions for:
  - Keywords (`const`, `from`, `func`, `op`, `type`, etc.)
  - Operators (`sort`, `where`, `yield`, `count`, `join`, etc.)
  - Functions (`abs`, `ceil`, `floor`, `len`, `split`, `upper`, etc.)
  - Aggregate functions (`count`, `sum`, `avg`, `max`, `min`, etc.)
  - Types (`int64`, `string`, `bool`, `time`, `duration`, etc.)

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

### VS Code

Add to your `settings.json`:

```json
{
  "superdb.lsp.path": "/path/to/superdb-lsp"
}
```

Or if using a generic LSP client extension:

```json
{
  "languageServerExample.trace.server": "verbose",
  "languageServerExample.serverPath": "/path/to/superdb-lsp"
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

### Server Capabilities

- **Text Document Sync**: Full document sync (mode 1)
- **Completion Provider**: Triggered by `.`, `|`, `(`, `:`, `=`

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
├── server_test.go   # Test harness
└── go.mod           # Go module definition
```

## SuperSQL Keywords Reference

### Keywords
`const`, `file`, `from`, `func`, `op`, `this`, `type`

### Operators
`assert`, `combine`, `cut`, `drop`, `file`, `fork`, `from`, `fuse`, `get`, `head`, `join`, `load`, `merge`, `over`, `pass`, `put`, `rename`, `sample`, `search`, `sort`, `summarize`, `switch`, `tail`, `top`, `uniq`, `where`, `yield`

### Functions
`abs`, `base64`, `bucket`, `cast`, `ceil`, `cidr_match`, `compare`, `coalesce`, `crop`, `error`, `every`, `fields`, `fill`, `flatten`, `floor`, `grep`, `grok`, `has`, `hex`, `has_error`, `is`, `is_error`, `join`, `kind`, `ksuid`, `len`, `levenshtein`, `log`, `lower`, `map`, `missing`, `nameof`, `nest_dotted`, `network_of`, `now`, `order`, `parse_uri`, `parse_zson`, `pow`, `quiet`, `regexp`, `regexp_replace`, `replace`, `round`, `rune_len`, `shape`, `split`, `sqrt`, `strftime`, `trim`, `typename`, `typeof`, `typeunder`, `under`, `unflatten`, `upper`

### Aggregates
`and`, `any`, `avg`, `collect`, `collect_map`, `count`, `dcount`, `fuse`, `max`, `min`, `or`, `sum`, `union`

### Types
`uint8`, `uint16`, `uint32`, `uint64`, `uint128`, `uint256`, `int8`, `int16`, `int32`, `int64`, `int128`, `int256`, `duration`, `time`, `float16`, `float32`, `float64`, `float128`, `float256`, `decimal32`, `decimal64`, `decimal128`, `decimal256`, `bool`, `bytes`, `string`, `ip`, `net`, `type`, `null`

## License

See the repository root for license information.
