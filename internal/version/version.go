package version

import (
	"fmt"
	"runtime"
)

// Build information. Will be set via ldflags during build.
var (
	Version   = "dev"      // Version string (git describe or tag)
	GitCommit = "unknown"  // Git commit hash  
	BuildTime = "unknown"  // Build timestamp
	GoVersion = runtime.Version()
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("%s (%s) built with %s at %s", Version, GitCommit, GoVersion, BuildTime)
}

// Short returns just the version string
func Short() string {
	return Version
}