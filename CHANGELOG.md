# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.1.0] - 2026-01-30

### Added
- Function: `concat` (concatenate multiple strings)
- TextMate: `values` operator to syntax highlighting

### Changed
- Version now follows brimdata/super release versions
- Updated parser API from `ParseQuery` to `Parse` (matches upstream refactor)
- Synced with brimdata/super v0.1.0 release (commit e8764da)

## [0.51231.1] - 2026-01-04

### Added
- Keywords: `filter`, `map`
- Aggregates: `first`, `last` (as aggregate functions)
- SQL type aliases: `char`, `cidr`, `double`, `float`, `inet`, `int`, `integer`, `interval`, `real`, `varchar`

### Changed
- `dcount()` return type from `uint64` to `int64` (matches upstream)
- Version now includes brimdata/super commit SHA as semver build metadata (e.g., `0.51231.0+5ea0cb5d`)

### Fixed
- Renamed TextMate bundle from `.tmb` to `.tmbundle` to fix "unknown format" error in editors

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
