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
	"os/exec"
	"strconv"
	"strings"

	"github.com/aliyun/git-repo-go/common"
)

// ExternalProtoHelper implements helper for unknown remote service.
type ExternalProtoHelper struct {
	sshInfo *SSHInfo

	program string
}

// NewExternalProtoHelper returns ExternalProtoHelper object.
func NewExternalProtoHelper(sshInfo *SSHInfo) *ExternalProtoHelper {
	if sshInfo.User == "" {
		sshInfo.User = "git"
	}
	return &ExternalProtoHelper{sshInfo: sshInfo}
}

// GetType returns remote server type.
func (v ExternalProtoHelper) GetType() string {
	return v.sshInfo.ProtoType
}

// GetSSHInfo returns SSHInfo object.
func (v ExternalProtoHelper) GetSSHInfo() *SSHInfo {
	return v.sshInfo
}

// Program is program name of remote helper.
func (v *ExternalProtoHelper) Program() string {
	if v.program == "" {
		v.program = "git-repo-helper-proto-" + strings.ToLower(v.sshInfo.ProtoType)
	}
	return v.program
}

// GetGitPushCommand reads upload options and returns git push command.
func (v ExternalProtoHelper) GetGitPushCommand(o *common.UploadOptions) (*GitPushCommand, error) {
	var (
		input   []byte
		output  []byte
		err     error
		pushCmd = GitPushCommand{}
	)

	input, err = json.Marshal(o)
	if err != nil {
		return nil, err
	}

	program, err := exec.LookPath(v.Program())
	if err != nil {
		return nil, fmt.Errorf("cannot find helper '%s'", v.Program())
	}

	cmdArgs := []string{program, "--upload"}
	if v.sshInfo.ProtoVersion > 0 {
		cmdArgs = append(cmdArgs, "--version", strconv.Itoa(v.sshInfo.ProtoVersion))
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = bytes.NewReader(input)
	output, err = cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fail to run %s: %s", v.Program(), exitError.Stderr)
		}
	}

	err = json.Unmarshal(output, &pushCmd)
	if err != nil {
		return nil, fmt.Errorf("invalid output from command '%s': %s", v.Program(), err)
	}
	return &pushCmd, nil
}

// GetDownloadRef returns reference name of the specific code review.
func (v ExternalProtoHelper) GetDownloadRef(cr, patch string) (string, error) {
	program, err := exec.LookPath(v.Program())
	if err != nil {
		return "", fmt.Errorf("cannot find helper '%s'", v.Program())
	}

	cmdArgs := []string{program, "--download"}
	if v.sshInfo.ProtoVersion > 0 {
		cmdArgs = append(cmdArgs, "--version", strconv.Itoa(v.sshInfo.ProtoVersion))
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = strings.NewReader(cr + " " + patch)
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("aaa fail to run %s: %s", v.Program(), exitError.Stderr)
		}
	}
	return strings.TrimSpace(string(out)), err
}
