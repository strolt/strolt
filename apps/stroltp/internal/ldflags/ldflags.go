// Package ldflags exposes build-time values injected via linker flags.
package ldflags

var (
	version    = "development"
	binaryName = "stroltp"
)

// GetVersion returns the build version.
func GetVersion() string {
	return version
}

// GetBinaryName returns the binary name.
func GetBinaryName() string {
	return binaryName
}
