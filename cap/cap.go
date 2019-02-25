package cap

import (
	"os"
	"runtime"

	"github.com/mattn/go-isatty"
)

// Symlink checks whether symlink is available for current system
func Symlink() bool {
	if runtime.GOOS == "windows" {
		return false
	}

	return true
}

// Isatty indicates current terminal is a interactive terminal
func Isatty() bool {
	if isatty.IsTerminal(os.Stdin.Fd()) &&
		isatty.IsTerminal(os.Stdout.Fd()) {
		return true
	} else if isatty.IsCygwinTerminal(os.Stdin.Fd()) &&
		isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return true
	}

	return false
}
