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

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/workspace"
	log "github.com/jiangxin/multi-log"
)

// WorkSpaceCommand implements load of workspace
type WorkSpaceCommand struct {
	ws  workspace.WorkSpace
	rws *workspace.RepoWorkSpace

	MirrorOK bool
	SingleOK bool
}

// WorkSpace loads workspace and return WorkSpace object
func (v *WorkSpaceCommand) WorkSpace() workspace.WorkSpace {
	var err error
	if v.ws == nil {
		v.ws, err = workspace.NewWorkSpace("")
		if err != nil {
			log.Fatal(err)
		}
	}
	if !v.SingleOK && config.IsSingleMode() {
		log.Fatal("cannot run in single mode")
	}
	if v.ws != nil {
		if !v.MirrorOK && v.ws.IsMirror() {
			log.Fatal("cannot run in a mirror")
		}
	}
	return v.ws
}

// RepoWorkSpace loads workspace and return RepoWorkSpace object
func (v *WorkSpaceCommand) RepoWorkSpace() *workspace.RepoWorkSpace {
	var err error
	if v.rws == nil {
		v.rws, err = workspace.NewRepoWorkSpace("")
		if err != nil {
			log.Fatal(err)
		}
	}
	if !v.SingleOK && config.IsSingleMode() {
		log.Fatal("cannot run in single mode")
	}
	if v.rws != nil {
		if !v.MirrorOK && v.rws.IsMirror() {
			log.Fatal("cannot run in a mirror")
		}
	}
	return v.rws
}

// ReloadRepoWorkSpace will reload workspace
func (v *WorkSpaceCommand) ReloadRepoWorkSpace() *workspace.RepoWorkSpace {
	v.ws = nil
	v.rws = nil
	return v.RepoWorkSpace()
}

// commandError is an error used to signal different error situations in command handling.
type commandError struct {
	s         string
	userError bool
}

func (c commandError) Error() string {
	return c.s
}

func (c commandError) isUserError() bool {
	return c.userError
}

func newUserError(a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: true}
}

func newUserErrorF(format string, a ...interface{}) commandError {
	return commandError{s: fmt.Sprintf(format, a...), userError: true}
}

func newSystemError(a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: false}
}

func newSystemErrorF(format string, a ...interface{}) commandError {
	return commandError{s: fmt.Sprintf(format, a...), userError: false}
}

// Catch some of the obvious user errors from Cobra.
// We don't want to show the usage message for every error.
// The below may be to generic. Time will show.
var userErrorRegexp = regexp.MustCompile("argument|flag|shorthand")

func isUserError(err error) bool {
	if cErr, ok := err.(commandError); ok && cErr.isUserError() {
		return true
	}

	return userErrorRegexp.MatchString(err.Error())
}

func min(args ...int) int {
	m := args[0]
	for _, arg := range args[1:] {
		if arg < m {
			m = arg
		}
	}
	return m
}

func userInput(prompt, defaultValue string) string {
	fmt.Print(prompt)

	if config.AssumeYes() {
		fmt.Println("Yes")
		return "yes"
	} else if config.AssumeNo() {
		fmt.Println("No")
		return "no"
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	if text == "" {
		return defaultValue
	}
	return text
}

func answerIsTrue(answer string) bool {
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "y" ||
		answer == "yes" ||
		answer == "t" ||
		answer == "true" ||
		answer == "on" ||
		answer == "1" {
		return true
	}
	return false
}
