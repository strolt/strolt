package ldflags

var (
	version    = "development"
	binaryName = "strolt"
)

func GetVersion() string {
	return version
}

func GetBinaryName() string {
	return binaryName
}
