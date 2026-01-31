package main

// Version follows brimdata/super release versions
// See: https://github.com/brimdata/super/releases
const Version = "0.1.0"

// SuperCommit is the brimdata/super commit SHA this version is synced to
// Updated by /sync command
const SuperCommit = "e8764da"

// FullVersion returns version with super commit as semver build metadata
func FullVersion() string {
	if SuperCommit != "" {
		return Version + "+" + SuperCommit
	}
	return Version
}
