// Package ldflags exposes build-time variables injected via -ldflags.
package ldflags

var (
	version    = "development"
	binaryName = "strolt"
)

// GetVersion returns the build version.
func GetVersion() string {
	return version
}

// GetBinaryName returns the binary name.
func GetBinaryName() string {
	return binaryName
}
