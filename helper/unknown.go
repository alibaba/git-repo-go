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

package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"code.alibaba-inc.com/force/git-repo/project"
)

// UnknownHelper implements helper for unknown remote service.
type UnknownHelper struct {
	RemoteType string
	program    string
}

// GetType returns remote server type.
func (v UnknownHelper) GetType() string {
	return v.RemoteType
}

// Program is program name of remote helper.
func (v *UnknownHelper) Program() string {
	if v.program == "" {
		v.program = "git-repo-helper-remote-" + strings.ToLower(v.RemoteType)
	}
	return v.program
}

// GetGitPushCommand reads upload options and returns git push command.
func (v UnknownHelper) GetGitPushCommand(o *project.UploadOptions) (*GitPushCommand, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)
	result, err := v.GetGitPushCommandPipe(reader)
	if err != nil {
		return nil, err
	}

	cmd := GitPushCommand{}
	err = json.Unmarshal(result, &cmd)
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

// GetGitPushCommand reads JSON from reader, and format it into proper JSON
// contains git push command.
func (v UnknownHelper) GetGitPushCommandPipe(reader io.Reader) ([]byte, error) {
	program, err := exec.LookPath(v.Program())
	if err != nil {
		return nil, fmt.Errorf("cannot find helper '%s'", v.Program())
	}

	cmd := exec.Command(program, "--upload")
	cmd.Stdin = reader
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fail to run %s: %s", v.Program(), exitError.Stderr)
		}
	}
	return bytes.TrimSpace(out), err
}

// GetDownloadRef returns reference name of the specific code review.
func (v UnknownHelper) GetDownloadRef(cr, patch string) (string, error) {
	program, err := exec.LookPath(v.Program())
	if err != nil {
		return "", fmt.Errorf("cannot find helper '%s'", v.Program())
	}

	cmd := exec.Command(program, "--download")
	cmd.Stdin = strings.NewReader(cr + " " + patch)
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("aaa fail to run %s: %s", v.Program(), exitError.Stderr)
		}
	}
	return strings.TrimSpace(string(out)), err
}
