package helper

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitSSHCmdEnv(t *testing.T) {
	assert := assert.New(t)
	os.Unsetenv("GIT_SSH_COMMAND")
	os.Unsetenv("GIT_SSH")
	os.Unsetenv("GIT_SSH_VARIANT")

	m := map[string]int{
		`"C:/Program Files/TortoiseGit/TortoisePlink.exe" arg1 arg2`: SSH_VARIANT_TORTOISEPLINK,
		"ssh":                         SSH_VARIANT_SSH,
		"ssh.exe -p 22 -o setenv=a=b": SSH_VARIANT_SSH,
		`"C:/Program Files/Plink/Plink.exe" arg1 arg2`: SSH_VARIANT_PLINK,
		"unknown program": SSH_VARIANT_SIMPLE,
	}

	for env, variant := range m {
		os.Setenv("GIT_SSH_COMMAND", env)
		cmd := NewSSHCmd()
		assert.Equal(variant, cmd.Variant(), fmt.Sprintf("fail on cmd: %s", env))
	}
}

func TestGitSSHEnv(t *testing.T) {
	assert := assert.New(t)
	os.Unsetenv("GIT_SSH_COMMAND")
	os.Unsetenv("GIT_SSH")
	os.Unsetenv("GIT_SSH_VARIANT")

	m := map[string]int{
		"ssh":                             SSH_VARIANT_SSH,
		"ssh.exe":                         SSH_VARIANT_SSH,
		"C:/ProgramFiles/Plink/Plink.exe": SSH_VARIANT_PLINK,
		"C:/Program Files/TortoiseGit/TortoisePlink.exe": SSH_VARIANT_TORTOISEPLINK,
		"unknown program": SSH_VARIANT_SIMPLE,
	}

	for env, variant := range m {
		os.Setenv("GIT_SSH", env)
		cmd := NewSSHCmd()
		assert.Equal(variant, cmd.Variant(), fmt.Sprintf("fail on cmd: %s", env))
	}
}

func TestGitSSHVariantEnv(t *testing.T) {
	assert := assert.New(t)
	os.Unsetenv("GIT_SSH_COMMAND")
	os.Unsetenv("GIT_SSH")
	os.Unsetenv("GIT_SSH_VARIANT")

	m := map[string]int{
		"auto":          SSH_VARIANT_SSH,
		"putty":         SSH_VARIANT_PUTTY,
		"plink":         SSH_VARIANT_PLINK,
		"tortoisePlink": SSH_VARIANT_TORTOISEPLINK,
		"simple":        SSH_VARIANT_SIMPLE,
		"ssh":           SSH_VARIANT_SSH,
		"unknown":       SSH_VARIANT_SSH,
	}

	for env, variant := range m {
		os.Setenv("GIT_SSH_VARIANT", env)
		cmd := NewSSHCmd()
		assert.Equal(variant, cmd.Variant(), fmt.Sprintf("fail on cmd: %s", env))
	}
}

func TestGitSSHCmd(t *testing.T) {
	var (
		host   = "example.com"
		port   = 22
		envs   = []string{"AGIT_FLOW=1", "GIT_PROTOCOL=2"}
		assert = assert.New(t)
	)

	os.Unsetenv("GIT_SSH_COMMAND")
	os.Unsetenv("GIT_SSH")
	os.Unsetenv("GIT_SSH_VARIANT")

	m := map[string][]string{
		"ssh -i ~/.ssh/mykey": []string{
			"ssh",
			"-i",
			"~/.ssh/mykey",
			"-o",
			"SendEnv=AGIT_FLOW",
			"-o",
			"SendEnv=GIT_PROTOCOL",
			"example.com",
		},
		"ssh": []string{
			"ssh",
			"-o",
			"SendEnv=AGIT_FLOW",
			"-o",
			"SendEnv=GIT_PROTOCOL",
			"example.com",
		},
		"\"C:/Program Files/TortoiseGit/TortoisePlink.exe\"": []string{
			"C:/Program Files/TortoiseGit/TortoisePlink.exe",
			"-batch",
			"example.com",
		},
		"\"C:/Program Files/Plink/plink.exe\" -some -options": []string{
			"C:/Program Files/Plink/plink.exe",
			"-some",
			"-options",
			"example.com",
		},
	}

	for key, value := range m {
		os.Setenv("GIT_SSH_COMMAND", key)
		cmd := NewSSHCmd()
		cmdArgs, newEnvs := cmd.Command(host, port, envs)
		assert.Equal(envs, newEnvs)
		assert.Equal(value, cmdArgs)
	}
}

func TestQuoteSSHCmd(t *testing.T) {
	var (
		host   = "example.com"
		port   = 22
		envs   = []string{"AGIT_FLOW=1", "GIT_PROTOCOL=2"}
		assert = assert.New(t)
	)

	os.Unsetenv("GIT_SSH_COMMAND")
	os.Unsetenv("GIT_SSH")
	os.Unsetenv("GIT_SSH_VARIANT")

	m := map[string]string{
		"ssh -i ~/.ssh/mykey": `ssh -i "~/.ssh/mykey" -o SendEnv=AGIT_FLOW -o SendEnv=GIT_PROTOCOL example.com`,
		"ssh":                 `ssh -o SendEnv=AGIT_FLOW -o SendEnv=GIT_PROTOCOL example.com`,

		"\"C:/Program Files/TortoiseGit/TortoisePlink.exe\"":  `"C:/Program Files/TortoiseGit/TortoisePlink.exe" -batch example.com`,
		"\"C:/Program Files/Plink/plink.exe\" -some -options": `"C:/Program Files/Plink/plink.exe" -some -options example.com`,
	}

	for key, value := range m {
		os.Setenv("GIT_SSH_COMMAND", key)
		cmd := NewSSHCmd()
		cmdArgs, newEnvs := cmd.Command(host, port, envs)
		shellCmd := NewShellCmdFromArgs(cmdArgs...)
		assert.Equal(envs, newEnvs)
		assert.Equal(value, shellCmd.QuoteCommand())
	}
}
