// Copyright © 2019 Alibaba Co. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package editor implements content edition.
package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"github.com/mattn/go-shellwords"
)

// theEditor is wapper for editor.
type theEditor struct {
	cfg    goconfig.GitConfig
	editor string
}

// Config returns git default settings.
func (v *theEditor) Config() goconfig.GitConfig {
	if v.cfg == nil {
		v.cfg = goconfig.DefaultConfig()
	}
	return v.cfg
}

// Editor returns program name of the editor.
func (v *theEditor) Editor() string {
	if v.editor == "" {
		v.editor = v.selectEditor()
	}
	return v.editor
}

func (v theEditor) selectEditor() string {
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
		if len(editor) > 0 && (editor[0] == '"' || editor[0] == '\'') {
			cmdArgs, err = shellwords.Parse(editor)

			if err != nil {
				log.Errorf("fail to parse editor '%s': %s", editor, err)
				cmdArgs = append(cmdArgs, editor)
			}
		} else {
			for i, c := range editor {
				if c == ' ' || c == '\t' {
					if path.Exist(editor[:i]) {
						cmdArgs = append(cmdArgs, editor[:i])
						args, err := shellwords.Parse(editor[i+1:])
						if err != nil {
							log.Errorf("fail to parse args'%s': %s", editor[i+1:], err)
							cmdArgs = append(cmdArgs, editor[i+1:])
						} else {
							cmdArgs = append(cmdArgs, args...)
						}
						break
					}
				}
			}
			if len(cmdArgs) == 0 {
				cmdArgs = append(cmdArgs, editor)
			}
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

// EditString starts the editor to edit data, and returns the edited data.
func (v theEditor) EditString(data string) string {
	var (
		err    error
		editor string
	)

	editor = v.Editor()
	if editor == ":" || !cap.Isatty() {
		if editor == ":" {
			log.Info("editor is ':', return directly")
		}
		log.Notef("no editor, input data unchanged")
		fmt.Println(data)
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

// EditString starts an editor to edit data, and returns the edited data.
func EditString(data string) string {
	e := theEditor{}
	return e.EditString(data)
}
