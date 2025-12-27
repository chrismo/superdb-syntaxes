# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.51224.2] - 2025-12-27

### Added
- SUP data file parsing with real syntax validation using `sup.Parser`
- SUP data file formatting with pretty-printing using `sup.Formatter`
- Diagnostics for invalid data syntax in `.sup` files
- Support for multiple values per data file
- Support for complex data types (arrays, nested records, timestamps, IPs, etc.)

### Changed
- `.sup` files now get parsed as data instead of being skipped entirely

## [0.51224.1] - 2025-12-24

### Added
- Patch version component to version format (0.YMMDD.P)

## [0.51224.0] - 2025-12-24

### Added
- Initial LSP implementation with diagnostics, completion, hover, signature help
- Document formatting for SuperSQL queries
- Skip diagnostics for `.sup` data files (treated as data, not queries)
- Golden tests for formatting
