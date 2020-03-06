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

	"github.com/alibaba/git-repo-go/cap"
	"github.com/alibaba/git-repo-go/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"github.com/mattn/go-shellwords"
)

var (
	cfg       goconfig.GitConfig
	editorCmd string
)

func config() goconfig.GitConfig {
	if cfg != nil {
		return cfg
	}

	cfg = goconfig.DefaultConfig()
	return cfg
}

// Editor returns program name of the editor.
func Editor() string {
	var (
		env string
	)

	if editorCmd != "" {
		return editorCmd
	}

	if env = os.Getenv("GIT_EDITOR"); env != "" {
		editorCmd = env
	} else if env = config().Get("core.editor"); env != "" {
		editorCmd = env
	} else if env = os.Getenv("VISUAL"); env != "" {
		editorCmd = env
	} else if env = os.Getenv("EDITOR"); env != "" {
		editorCmd = env
	} else if os.Getenv("TERM") == "dumb" {
		log.Fatal(
			"No editor specified in GIT_EDITOR, core.editor, VISUAL or EDITOR.\n" +
				"Tried to fall back to vi but terminal is dumb.  Please configure at\n" +
				"least one of these before using this command.")
	} else {
		for _, c := range []string{"vim", "vi", "emacs", "nano"} {
			if path, err := exec.LookPath(c); err == nil {
				editorCmd = path
				break
			}
		}
	}
	return editorCmd
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

// EditString starts an editor to edit data, and returns the edited data.
func EditString(data string) string {
	var (
		err    error
		editor string
	)

	editor = Editor()
	if editor == ":" || editor == "" || !cap.Isatty() {
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
