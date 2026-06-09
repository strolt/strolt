// Package ldflags exposes build-time variables injected via linker flags.
package ldflags

var version = "development"

// GetVersion returns the application version set at build time.
func GetVersion() string {
	return version
}
