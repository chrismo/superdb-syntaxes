package main

// Version information - update on each sync with upstream
// Format: 0.YMMDD.P where Y=last digit of year, MM=month, DD=day, P=patch
// Based on latest brimdata/super PEG parser commit date
// Patch number increments for internal changes between syncs
const Version = "0.60130.0"

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
