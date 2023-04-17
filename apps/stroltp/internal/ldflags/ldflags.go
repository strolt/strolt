package ldflags

var (
	version    = "development"
	binaryName = "stroltp"
)

func GetVersion() string {
	return version
}

func GetBinaryName() string {
	return binaryName
}
