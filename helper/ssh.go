package helper

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/git-repo-go/config"
	log "github.com/jiangxin/multi-log"
)

// Define constants for SSH variant types.
const (
	SSHVariantAuto = iota
	SSHVariantSimple
	SSHVariantSSH
	SSHVariantPlink
	SSHVariantPutty
	SSHVariantTortoisePlink
)

const (
	sshVariantDetectTimeout = 2
)

// SSHCmd is composor for SSH command.
type SSHCmd struct {
	ssh     string
	args    []string
	variant int
}

// NewSSHCmd returns SSHCmd by inspecting environments like `GIT_SSH_COMMAND`.
func NewSSHCmd() *SSHCmd {
	var (
		ssh  string
		args []string
	)
	cmd := os.Getenv("GIT_SSH_COMMAND")
	if cmd == "" {
		cmd = config.GitDefaultConfig.Get("core.sshcommand")
	}
	if cmd != "" {
		shellCmd := NewShellCmd(cmd, true)
		ssh = shellCmd.Cmd
		args = shellCmd.Args
	} else {
		ssh = os.Getenv("GIT_SSH")
	}
	if ssh == "" {
		ssh = "ssh"
	}
	return &SSHCmd{ssh: ssh, args: args}
}

// SSH returns ssh program name.
func (v *SSHCmd) SSH() string {
	return v.ssh
}

// Args returns default ssh arguments.
func (v *SSHCmd) Args() []string {
	return v.args
}

// Variant indicates ssh variant type.
func (v *SSHCmd) Variant() int {
	if v.variant > 0 {
		return v.variant
	}
	setting := os.Getenv("GIT_SSH_VARIANT")
	if setting == "" {
		setting = config.GitDefaultConfig.Get("ssh.variant")
	}
	if setting != "" {
		switch strings.ToLower(setting) {
		case "auto":
			v.variant = SSHVariantAuto
		case "plink":
			v.variant = SSHVariantPlink
		case "putty":
			v.variant = SSHVariantPutty
		case "tortoiseplink":
			v.variant = SSHVariantTortoisePlink
		case "simple":
			v.variant = SSHVariantSimple
		default:
			v.variant = SSHVariantSSH
		}
	} else {
		switch strings.ToLower(path.Base(v.SSH())) {
		case "ssh", "ssh.exe":
			v.variant = SSHVariantSSH
		case "plink", "plink.exe":
			v.variant = SSHVariantPlink
		case "tortoiseplink", "tortoiseplink.exe":
			v.variant = SSHVariantTortoisePlink
		default:
			v.variant = SSHVariantAuto
		}
	}

	if v.variant == SSHVariantAuto {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			sshVariantDetectTimeout*time.Second,
		)
		defer cancel()
		if err := exec.CommandContext(ctx, v.SSH(), "-G", "127.0.0.1").Run(); err != nil {
			v.variant = SSHVariantSimple
		} else {
			v.variant = SSHVariantSSH
		}
	}
	return v.variant
}

// Command returns command and environments used for ssh connection.
func (v *SSHCmd) Command(host string, port int, envs []string) ([]string, []string) {
	cmdArgs := []string{v.SSH()}
	cmdArgs = append(cmdArgs, v.Args()...)
	if v.Variant() == SSHVariantSSH {
		for _, env := range envs {
			cmdArgs = append(cmdArgs, "-o", "SendEnv="+strings.Split(env, "=")[0])
		}
	}
	if v.Variant() == SSHVariantTortoisePlink {
		cmdArgs = append(cmdArgs, "-batch")
	}
	if port > 0 && port != 22 {
		switch v.Variant() {
		case SSHVariantSSH:
			cmdArgs = append(cmdArgs, "-p", strconv.Itoa(port))
		case SSHVariantPutty, SSHVariantPlink, SSHVariantTortoisePlink:
			cmdArgs = append(cmdArgs, "-P", strconv.Itoa(port))
		case SSHVariantSimple:
			log.Fatal("ssh variant 'simple' does not support setting port")
		}
	}

	if host != "" {
		cmdArgs = append(cmdArgs, host)
	}
	return cmdArgs, envs
}
