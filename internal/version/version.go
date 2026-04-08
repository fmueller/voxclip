package version

// Version is overridden at build time via ldflags.
var Version = "1.1.0"

// Resolve returns the version string.
func Resolve() string {
	if Version == "" {
		return "0.0.0"
	}
	return Version
}
