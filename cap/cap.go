package cap

import (
	"fmt"
	"runtime"
)

// Capabilities is used to check OS capabilities
type Capabilities struct {
}

// Cap is instance of Capabilities
var (
	Cap *Capabilities
)

// Symlink checks whether symlink is available for current system
func (v *Capabilities) Symlink() bool {
	if runtime.GOOS == "windows" {
		return false
	}

	fmt.Printf("os: %s\n", runtime.GOOS)
	return true
}

func init() {
	Cap = &Capabilities{}
}
