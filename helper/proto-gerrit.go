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
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/git-repo-go/common"
	"github.com/alibaba/git-repo-go/config"
	log "github.com/jiangxin/multi-log"
)

// GerritProtoHelper wraps helper for gerrit server.
type GerritProtoHelper struct {
	sshInfo *SSHInfo
}

// NewGerritProtoHelper returns GerritProtoHelper object.
func NewGerritProtoHelper(sshInfo *SSHInfo) *GerritProtoHelper {
	if sshInfo.User == "" {
		sshInfo.User = "<email>"
	}
	if sshInfo.ReviewRefPattern == "" {
		sshInfo.ReviewRefPattern = "refs/changes/{id:right:2}/{id}/{patch}"
	}
	return &GerritProtoHelper{sshInfo: sshInfo}
}

// GetType returns remote server type.
func (v GerritProtoHelper) GetType() string {
	return ProtoTypeGerrit
}

// GetSSHInfo returns SSHInfo object.
func (v GerritProtoHelper) GetSSHInfo() *SSHInfo {
	return v.sshInfo
}

// GetGitPushCommand reads upload options and returns git push command.
func (v GerritProtoHelper) GetGitPushCommand(o *common.UploadOptions) (*GitPushCommand, error) {
	if !o.CodeReview.Empty() {
		return nil, fmt.Errorf("Change code review by ID is not allowed in Gerrit")
	}

	cmds := []string{"git", "push"}

	if o.RemoteURL == "" {
		return nil, errors.New("empty review url for helper")
	}
	gitURL := config.ParseGitURL(o.RemoteURL)
	if gitURL == nil || (!gitURL.IsSSH() && !gitURL.IsHTTP()) {
		return nil, fmt.Errorf("bad review url: %s", o.RemoteURL)
	}

	if gitURL.IsSSH() {
		cmds = append(cmds, "--receive-pack=gerrit receive-pack")
	}
	for _, pushOption := range o.PushOptions {
		cmds = append(cmds, "-o", pushOption)
	}
	if o.RemoteName != "" {
		cmds = append(cmds, o.RemoteName)
	} else {
		cmds = append(cmds, o.RemoteURL)
	}

	destBranch := o.DestBranch
	if strings.HasPrefix(destBranch, config.RefsHeads) {
		destBranch = strings.TrimPrefix(destBranch, config.RefsHeads)
	}
	if destBranch == "" {
		return nil, errors.New("empty dest branch for helper")
	}

	uploadType := "for"
	refSpec := ""
	if o.Draft {
		uploadType = "drafts"
	}

	localBranch := o.LocalBranch
	if strings.HasPrefix(localBranch, config.RefsHeads) {
		localBranch = strings.TrimPrefix(localBranch, config.RefsHeads)
	}
	if localBranch == "" {
		refSpec = "HEAD"
	} else {
		refSpec = config.RefsHeads + localBranch
	}

	refSpec += fmt.Sprintf(":refs/%s/%s",
		uploadType,
		destBranch)

	if o.AutoTopic && localBranch != "" {
		refSpec = refSpec + "/" + localBranch
	}

	opts := []string{}
	if o.People != nil && len(o.People) > 0 {
		for _, u := range o.People[0] {
			opts = append(opts, "r="+u)
		}
	}
	if o.People != nil && len(o.People) > 1 {
		for _, u := range o.People[1] {
			opts = append(opts, "cc="+u)
		}
	}
	if o.NoEmails {
		opts = append(opts, "notify=NONE")
	}
	if o.Private {
		opts = append(opts, "private")
	}
	if o.WIP {
		opts = append(opts, "wip")
	}
	if len(opts) > 0 {
		refSpec = refSpec + "%" + strings.Join(opts, ",")
	}

	cmds = append(cmds, refSpec)

	cmd := GitPushCommand{}
	cmd.Cmd = cmds[0]
	cmd.Args = cmds[1:]
	return &cmd, nil
}

// GetDownloadRef returns reference name of the specific code review.
func (v GerritProtoHelper) GetDownloadRef(cr, patch string) (string, error) {
	var (
		patchID int
		err     error
	)

	_, err = strconv.Atoi(cr)
	if err != nil {
		return "", fmt.Errorf("bad review ID %s: %s", cr, err)
	}

	if patch != "" {
		patchID, err = strconv.Atoi(patch)
		if err != nil {
			return "", fmt.Errorf("bad patch ID %s: %s", patch, err)
		}
	}

	if patchID == 0 {
		log.Warn("Patch ID should not be 0, set it to 1")
		patch = "1"
	}
	return v.sshInfo.GetReviewRef(cr, patch)
}
