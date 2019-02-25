package project

import (
	"os"
	"os/exec"
	"strings"

	"github.com/jiangxin/multi-log"
)

func executeCommand(args ...string) error {
	return executeCommandIn("", args)
}

func executeCommandIn(cwd string, args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	if cwd != "" {
		if _, err := os.Stat(cwd); err != nil {
			log.Errorf("cannot enter '%s' to run %s",
				cwd,
				strings.Join(args, " "))
		}
		cmd.Dir = cwd
	}
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
