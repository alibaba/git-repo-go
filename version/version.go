package version

var (
	// Version of git-repo
	Version = "undefined"
)

// GetVersion show git-repo version
func GetVersion() string {
	return Version
}
