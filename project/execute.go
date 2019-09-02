package project

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/jiangxin/multi-log"
)

// CmdExecResult holds command output, and error.
type CmdExecResult struct {
	Project *Project
	Out     []byte
	Error   error
}

// NewCmdExecResult creates new instance of CmdExecResult.
func NewCmdExecResult(p *Project) *CmdExecResult {
	result := CmdExecResult{
		Project: p,
	}
	return &result
}

// Stdout is command output on stdout.
func (v *CmdExecResult) Stdout() string {
	return string(v.Out)
}

// Stderr is command output on stderr.
func (v *CmdExecResult) Stderr() string {
	if v.Error == nil {
		return ""
	}

	if exitError, ok := v.Error.(*exec.ExitError); ok {
		return string(exitError.Stderr)
	}

	return v.Error.Error()
}

// Empty indicates output and error output is empty.
func (v *CmdExecResult) Empty() bool {
	return len(v.Stdout()) == 0 && len(v.Stderr()) == 0
}

// Success indicates command runs successfully or not.
func (v CmdExecResult) Success() bool {
	if v.Error == nil {
		return true
	}

	if exitError, ok := v.Error.(*exec.ExitError); ok {
		return exitError.Success()
	}

	return false
}

// ExecuteCommand runs command.
func (v Project) ExecuteCommand(args ...string) *CmdExecResult {
	result := CmdExecResult{
		Project: &v,
	}
	cmd := exec.Command(args[0], args[1:]...)
	if v.IsMirror() {
		cmd.Dir = v.GitDir
	} else {
		cmd.Dir = v.WorkDir
	}
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
