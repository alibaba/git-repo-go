package editor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"github.com/mattn/go-shellwords"
)

// Editor is used to edit file
type Editor struct {
	cfg    goconfig.GitConfig
	editor string
}

// Config returns git default settings
func (v *Editor) Config() goconfig.GitConfig {
	if v.cfg == nil {
		v.cfg = goconfig.DefaultConfig()
	}
	return v.cfg
}

// Editor returns editor program name
func (v *Editor) Editor() string {
	if v.editor == "" {
		v.editor = v.selectEditor()
	}
	return v.editor
}

func (v Editor) selectEditor() string {
	var e string

	e = os.Getenv("GIT_EDITOR")
	if e != "" {
		return e
	}

	e = v.Config().Get("core.editor")
	if e != "" {
		return e
	}

	e = os.Getenv("VISUAL")
	if e != "" {
		return e
	}

	e = os.Getenv("EDITOR")
	if e != "" {
		return e
	}

	if os.Getenv("TERM") == "dumb" {
		log.Fatal(
			"No editor specified in GIT_EDITOR, core.editor, VISUAL or EDITOR.\n" +
				"Tried to fall back to vi but terminal is dumb.  Please configure at\n" +
				"least one of these before using this command.")

	}

	return "vi"
}

func editorCommands(editor string, args ...string) []string {
	var (
		cmdArgs = []string{}
		err     error
	)

	if cap.IsWindows() {
		// Split on spaces, respecting quoted strings
		cmdArgs, err = shellwords.Parse(editor)
		if err != nil {
			log.Errorf("fail to parse editor '%s': %s", editor, err)
		}
	} else if regexp.MustCompile(`^.*[$ \t'].*$`).MatchString(editor) {
		// See: https://gerrit-review.googlesource.com/c/git-repo/+/16156
		cmdArgs = append(cmdArgs,
			"sh",
			"-c",
			editor+` "$@"`,
			"sh")
	} else {
		cmdArgs = append(cmdArgs, editor)
	}
	cmdArgs = append(cmdArgs, args...)
	return cmdArgs
}

// EditString starts editor and returns data after edition
func (v Editor) EditString(data string) string {
	var (
		err    error
		editor string
	)

	editor = v.Editor()
	if editor == ":" {
		log.Infof("editor is ':', return directly:\n%s", data)
		return data
	}

	tmpfile, err := ioutil.TempFile("", "go-repo-edit-*")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		log.Fatal(err)
	}

	cmdArgs := editorCommands(editor, tmpfile.Name())
	if cap.Isatty() {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Errorf("fail to run '%s' to edit script: %s",
				strings.Join(cmdArgs, " "),
				err)
		}
	} else {
		log.Notef("will execute: %s", strings.Join(cmdArgs, " "))
		log.Debugf("data ready for editor:\n%s", data)
	}

	f, err := os.Open(tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return string(buf)
}
