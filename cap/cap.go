package cap

import (
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
)

// WindowsInterface is interface to check OS type
type WindowsInterface interface {
	IsWindows() bool
}

// TTYInterface is interface to check terminal is a tty
type TTYInterface interface {
	Isatty() bool
}

// Export interfaces, use can override these interfaces by mocking
var (
	CapWindows WindowsInterface
	CapTTY     TTYInterface
)

// Windows implements WindowsInterface
type Windows struct {
}

// IsWindows returns true if current OS is Windows
func (v Windows) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// TTY implements TTYInterface
type TTY struct {
}

// Isatty is true if has terminal
func (v TTY) Isatty() bool {
	if isatty.IsTerminal(os.Stdin.Fd()) &&
		isatty.IsTerminal(os.Stdout.Fd()) {
		return true
	} else if isatty.IsCygwinTerminal(os.Stdin.Fd()) &&
		isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return true
	}

	return false
}

// IsWindows checks whether current OS is windows
func IsWindows() bool {
	return CapWindows.IsWindows()
}

// Symlink checks whether symlink is available for current system
func Symlink() bool {
	return !IsWindows()
}

// Isatty indicates current terminal is a interactive terminal
func Isatty() bool {
	return CapTTY.Isatty()
}

func init() {
	CapWindows = &Windows{}
	CapTTY = &TTY{}
}
