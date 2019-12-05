package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellCmd(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Command  string
		ShellCmd ShellCmd
	}{
		{
			"ssh -p 10022 -o SetEnv='AGIT_FLOW=1' hostname command",
			ShellCmd{
				Cmd: "ssh",
				Args: []string{
					"-p",
					"10022",
					"-o",
					"SetEnv=AGIT_FLOW=1",
					"hostname",
					"command",
				}},
		},
		{
			"ssh   -p     10022	 -o	 SetEnv='AGIT_FLOW=1'  hostname command",
			ShellCmd{
				Cmd: "ssh",
				Args: []string{
					"-p",
					"10022",
					"-o",
					"SetEnv=AGIT_FLOW=1",
					"hostname",
					"command",
				}},
		},
		{
			"ssh",
			ShellCmd{
				Cmd:  "ssh",
				Args: nil,
			},
		},
		{
			"'s'\"s\"'h'",
			ShellCmd{
				Cmd:  "ssh",
				Args: nil,
			},
		},
		{
			`"C:\Program Files\ssh.exe" 'host name' `,
			ShellCmd{
				Cmd: "C:\\Program Files\\ssh.exe",
				Args: []string{
					"host name",
				}},
		},
		{
			"\"C:\\Program Files\\ssh.exe\" \"host name\"",
			ShellCmd{
				Cmd: `C:\Program Files\ssh.exe`,
				Args: []string{
					"host name",
				}},
		},
	}

	for _, testCase := range testCases {
		cmd := NewShellCmd(testCase.Command, true)
		assert.Equal(testCase.ShellCmd, *cmd)
	}
}

func TestShellCmdQuote(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		Command  string
		QuoteCmd string
	}{
		{
			"ssh -p 10022 -o SetEnv='AGIT_FLOW=1' hostname command",
			"ssh -p 10022 -o SetEnv=AGIT_FLOW=1 hostname command",
		},
		{
			"ssh -p 10022 -o SetEnv=AGIT_FLOW=1 hostname command",
			"ssh -p 10022 -o SetEnv=AGIT_FLOW=1 hostname command",
		},
		{
			"'s'\"s\"'h'",
			"ssh",
		},
		{
			`"C:/Program Files/ssh.exe" 'host name1' `,
			`"C:/Program Files/ssh.exe" "host name1"`,
		},
		{
			`"C:\Program Files\ssh.exe" 'host name2' `,
			"\"C:\\Program Files\\ssh.exe\" \"host name2\"",
		},
		{
			"\"C:\\Program Files\\ssh.exe\" \"host name3\"",
			"\"C:\\Program Files\\ssh.exe\" \"host name3\"",
		},
	}

	for _, testCase := range testCases {
		cmd := NewShellCmd(testCase.Command, true)
		assert.Equal(testCase.QuoteCmd, cmd.QuoteCommand())
	}
}
