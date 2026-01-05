# TODO

## Release Artifacts

- [ ] Add version to TextMate bundle filename (`spq-textmate-bundle-v0.51224.2.zip`)
- [ ] Create a `.sup` TextMate grammar for data file syntax highlighting
- [ ] Create separate VS Code extension repo (like IntelliJ plugin repo) that packages LSP + grammars

## Editor Documentation

Add setup instructions to README for:

- [ ] Neovim (lspconfig)
- [ ] Vim (vim-lsp or coc.nvim)
- [ ] Emacs (lsp-mode or eglot)
- [ ] Zed
- [ ] Helix
- [ ] Sublime Text

## Repository Rename

~~Consider renaming from `superdb-syntaxes` to `superdb-lsp`~~ ✓ Done

- [x] Decide on rename (recommendation: `superdb-lsp`)
- [x] Update internal references
- [ ] Update GitHub repo name (manual step on GitHub)

## Format 

- [ ] the snapshot tests don't fail if I change the expected. 
- [ ] added a .sup format test, also not failing
- [ ] rename Golden -> Snapshot in any of the codebase.