package helper

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"code.alibaba-inc.com/force/git-repo/config"
	log "github.com/jiangxin/multi-log"
)

const (
	SSH_VARIANT_AUTO = iota
	SSH_VARIANT_SIMPLE
	SSH_VARIANT_SSH
	SSH_VARIANT_PLINK
	SSH_VARIANT_PUTTY
	SSH_VARIANT_TORTOISEPLINK
)

const (
	sshVariantDetectTimeout = 2
)

type SSHCmd struct {
	ssh     string
	args    []string
	variant int
}

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
			v.variant = SSH_VARIANT_AUTO
		case "plink":
			v.variant = SSH_VARIANT_PLINK
		case "putty":
			v.variant = SSH_VARIANT_PUTTY
		case "tortoiseplink":
			v.variant = SSH_VARIANT_TORTOISEPLINK
		case "simple":
			v.variant = SSH_VARIANT_SIMPLE
		default:
			v.variant = SSH_VARIANT_SSH
		}
	} else {
		switch strings.ToLower(path.Base(v.SSH())) {
		case "ssh", "ssh.exe":
			v.variant = SSH_VARIANT_SSH
		case "plink", "plink.exe":
			v.variant = SSH_VARIANT_PLINK
		case "tortoiseplink", "tortoiseplink.exe":
			v.variant = SSH_VARIANT_TORTOISEPLINK
		default:
			v.variant = SSH_VARIANT_AUTO
		}
	}

	if v.variant == SSH_VARIANT_AUTO {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			sshVariantDetectTimeout*time.Second,
		)
		defer cancel()
		if err := exec.CommandContext(ctx, v.SSH(), "-G", "127.0.0.1").Run(); err != nil {
			v.variant = SSH_VARIANT_SIMPLE
		} else {
			v.variant = SSH_VARIANT_SSH
		}
	}
	return v.variant
}

// Commands returns command and environments.
func (v *SSHCmd) Command(host string, port int, envs []string) ([]string, []string) {
	cmdArgs := []string{v.SSH()}
	cmdArgs = append(cmdArgs, v.Args()...)
	if v.Variant() == SSH_VARIANT_SSH {
		for _, env := range envs {
			cmdArgs = append(cmdArgs, "-o", "SendEnv="+strings.Split(env, "=")[0])
		}
	}
	if v.Variant() == SSH_VARIANT_TORTOISEPLINK {
		cmdArgs = append(cmdArgs, "-batch")
	}
	if port > 0 && port != 22 {
		switch v.Variant() {
		case SSH_VARIANT_SSH:
			cmdArgs = append(cmdArgs, "-p", strconv.Itoa(port))
		case SSH_VARIANT_PUTTY, SSH_VARIANT_PLINK, SSH_VARIANT_TORTOISEPLINK:
			cmdArgs = append(cmdArgs, "-P", strconv.Itoa(port))
		case SSH_VARIANT_SIMPLE:
			log.Fatal("ssh variant 'simple' does not support setting port")
		}
	}

	if host != "" {
		cmdArgs = append(cmdArgs, host)
	}
	return cmdArgs, envs
}
