package project

import (
	"os"
	"os/exec"
	"strings"

	"github.com/jiangxin/multi-log"
)

// CmdExecResult
type CmdExecResult struct {
	Project *Project
	Out     []byte
	Error   error
}

func (v *CmdExecResult) Stdout() string {
	return string(v.Out)
}

func (v *CmdExecResult) Stderr() string {
	if v.Error == nil {
		return ""
	}

	if exitError, ok := v.Error.(*exec.ExitError); ok {
		return string(exitError.Stderr)
	}

	return v.Error.Error()
}

func (v CmdExecResult) Success() bool {
	if v.Error == nil {
		return true
	}

	if exitError, ok := v.Error.(*exec.ExitError); ok {
		return exitError.Success()
	}

	return false
}

func (v Project) ExecuteCommand(args ...string) *CmdExecResult {
	result := CmdExecResult{
		Project: &v,
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = v.WorkDir
	cmd.Stdin = nil
	result.Out, result.Error = cmd.Output()
	return &result
}

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
