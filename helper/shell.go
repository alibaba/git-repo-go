package helper

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

var (
	NormalArgsPattern = regexp.MustCompile(`^[0-9a-zA-Z:/%,.@+=_-]+$`)
)

type ShellCmd struct {
	Cmd  string
	Args []string
}

func (v ShellCmd) QuoteCommand() string {
	cmd := bytes.NewBuffer([]byte{})
	cmd.WriteString(v.quoteString(v.Cmd))
	for _, arg := range v.Args {
		cmd.WriteByte(' ')
		cmd.WriteString(v.quoteString(arg))
	}
	return cmd.String()
}

// quoteString adds quote to string, not convert so many chars like `strconv.Quote`.
func (v ShellCmd) quoteString(s string) string {
	var (
		hasBackSlash bool
	)

	if (s[0] == '"' || s[0] == '\'') && s[0] == s[len(s)-1] {
		return s
	}
	if NormalArgsPattern.MatchString(s) {
		return s
	}

	buf := bytes.NewBuffer([]byte{'"'})
	for _, c := range s {
		if hasBackSlash {
			hasBackSlash = false
			buf.WriteRune(c)
			continue
		}
		if c == '\\' {
			hasBackSlash = true
			buf.WriteRune(c)
			continue
		}
		switch c {
		case '"':
			buf.WriteString("\\\"")
		default:
			buf.WriteRune(c)
		}
	}
	buf.WriteRune('"')
	return buf.String()
}

func NewShellCmdFromArgs(args ...string) *ShellCmd {
	shellCmd := ShellCmd{}
	shellCmd.Cmd = args[0]
	if len(args) > 1 {
		shellCmd.Args = args[1:]
	}
	return &shellCmd
}

func NewShellCmd(cmd string, withArgs bool) *ShellCmd {
	shellCmd := ShellCmd{}
	cmd = strings.TrimSpace(cmd)
	if !withArgs {
		shellCmd.Cmd = cmd
		return &shellCmd
	}

	var (
		quoteChar rune
		isQuote   = false
		isSpace   = false
		token     = make([]rune, 0, 64)
		cmdArgs   []string
	)
	for _, c := range cmd {
		if !isQuote && (c == '\'' || c == '"') {
			quoteChar = c
			isQuote = true
			isSpace = false
			continue
		}
		if isQuote {
			if c == quoteChar {
				isQuote = false
				isSpace = false
				continue
			}
			token = append(token, c)
			continue
		}
		if unicode.IsSpace(c) {
			if isSpace {
				continue
			}
			isSpace = true
			cmdArgs = append(cmdArgs, string(token))
			token = token[0:0]
			continue
		}
		isSpace = false
		token = append(token, c)
	}
	if len(token) > 0 {
		cmdArgs = append(cmdArgs, string(token))
	}
	if len(cmdArgs) > 0 {
		shellCmd.Cmd = cmdArgs[0]
	}
	if len(cmdArgs) > 1 {
		shellCmd.Args = cmdArgs[1:]
	}
	return &shellCmd
}
