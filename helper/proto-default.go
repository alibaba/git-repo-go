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
	"errors"

	"github.com/aliyun/git-repo-go/common"
)

// DefaultProtoHelper implements helper for unknown remote service.
type DefaultProtoHelper struct {
	sshInfo *SSHInfo
}

// NewDefaultProtoHelper returns DefaultProtoHelper object.
func NewDefaultProtoHelper(sshInfo *SSHInfo) *DefaultProtoHelper {
	h := DefaultProtoHelper{sshInfo: sshInfo}
	return &h
}

// GetType returns remote server type.
func (v DefaultProtoHelper) GetType() string {
	return ""
}

// GetSSHInfo returns SSHInfo object.
func (v DefaultProtoHelper) GetSSHInfo() *SSHInfo {
	return v.sshInfo
}

// GetGitPushCommand reads upload options and returns git push command.
func (v DefaultProtoHelper) GetGitPushCommand(o *common.UploadOptions) (*GitPushCommand, error) {
	return nil, errors.New("not implement")
}

// GetDownloadRef returns reference name of the specific code review.
func (v DefaultProtoHelper) GetDownloadRef(cr, patch string) (string, error) {
	return "", errors.New("not implement")
}
