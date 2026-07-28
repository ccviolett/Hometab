package buildinfo

import "runtime"

// Injected at build time via -ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func GoVersion() string {
	return runtime.Version()
}
