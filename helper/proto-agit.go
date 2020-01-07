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

	"github.com/aliyun/git-repo-go/cap"
	"github.com/aliyun/git-repo-go/common"
	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/encode"
	"github.com/aliyun/git-repo-go/version"
	log "github.com/jiangxin/multi-log"
)

// AGitProtoHelper implements helper for AGit server.
type AGitProtoHelper struct {
	sshInfo *SSHInfo
}

// NewAGitProtoHelper returns AGitProtoHelper object.
func NewAGitProtoHelper(sshInfo *SSHInfo) *AGitProtoHelper {
	sshInfo.User = "git"
	return &AGitProtoHelper{sshInfo: sshInfo}
}

// GetType returns remote server type.
func (v AGitProtoHelper) GetType() string {
	return ProtoTypeAGit
}

// GetSSHInfo returns SSHInfo object.
func (v AGitProtoHelper) GetSSHInfo() *SSHInfo {
	return v.sshInfo
}

// GetGitPushCommand reads upload options and returns git push command.
func (v AGitProtoHelper) GetGitPushCommand(o *common.UploadOptions) (*GitPushCommand, error) {
	var (
		gitPushCmd = GitPushCommand{}
	)

	cmds := []string{"git", "push"}

	if o.RemoteURL == "" {
		return nil, errors.New("empty review url for helper")
	}
	gitURL := config.ParseGitURL(o.RemoteURL)
	if gitURL == nil || (gitURL.Proto != "ssh" && gitURL.Proto != "http" && gitURL.Proto != "https") {
		return nil, fmt.Errorf("bad review URL: %s", o.RemoteURL)
	}

	gitRepoAgent := "git-repo" + "/" + version.GetVersion()
	if gitURL.IsSSH() {
		switch v.sshInfo.ProtoVersion {
		case 0:
			cmds = append(cmds, "--receive-pack=agit-receive-pack")
		default:
			gitPushCmd.Env = []string{
				"AGIT_FLOW=" + gitRepoAgent,
			}
		}
	} else {
		gitPushCmd.GitConfig = []string{
			fmt.Sprintf(`http.extraHeader=AGIT-FLOW: %s`, gitRepoAgent),
		}
	}

	gitCanPushOptions := cap.GitCanPushOptions()
	if len(o.PushOptions) > 0 {
		if !gitCanPushOptions {
			log.Warnf("cannot send push options, for your git version is too low")
		} else {
			for _, pushOption := range o.PushOptions {
				cmds = append(cmds, "-o", pushOption)
			}
		}
	}

	uploadType := ""
	refSpec := ""
	localBranch := o.LocalBranch
	if strings.HasPrefix(localBranch, config.RefsHeads) {
		localBranch = strings.TrimPrefix(localBranch, config.RefsHeads)
	}
	if localBranch == "" {
		refSpec = "HEAD"
	} else {
		refSpec = config.RefsHeads + localBranch
	}

	if !o.CodeReview.Empty() {
		uploadType = "for-review"
		refSpec += fmt.Sprintf(":refs/%s/%s",
			uploadType,
			o.CodeReview.ID)
	} else {
		if o.Draft {
			uploadType = "drafts"
		} else {
			uploadType = "for"
		}

		destBranch := o.DestBranch
		if strings.HasPrefix(destBranch, config.RefsHeads) {
			destBranch = strings.TrimPrefix(destBranch, config.RefsHeads)
		}

		refSpec += fmt.Sprintf(":refs/%s/%s/%s",
			uploadType,
			destBranch,
			localBranch)
	}

	if gitCanPushOptions {
		if o.Title != "" {
			cmds = append(cmds, "-o", "title="+encode.B64Encode(o.Title))
		}
		if o.Description != "" {
			cmds = append(cmds, "-o", "description="+encode.B64Encode(o.Description))
		}
		if o.Issue != "" {
			cmds = append(cmds, "-o", "issue="+encode.B64Encode(o.Issue))
		}
		if o.People != nil && len(o.People) > 0 && len(o.People[0]) > 0 {
			reviewers := strings.Join(o.People[0], ",")
			cmds = append(cmds, "-o", "reviewers="+encode.B64Encode(reviewers))
		}
		if o.People != nil && len(o.People) > 1 && len(o.People[1]) > 0 {
			cc := strings.Join(o.People[1], ",")
			cmds = append(cmds, "-o", "cc="+encode.B64Encode(cc))
		}

		if o.NoEmails {
			cmds = append(cmds, "-o", "notify=no")
		}
		if o.Private {
			cmds = append(cmds, "-o", "private=yes")
		}
		if o.WIP {
			cmds = append(cmds, "-o", "wip=yes")
		}
		if o.OldOid != "" {
			cmds = append(cmds, "-o", "oldoid="+o.OldOid)
		}
	} else {
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
		if o.OldOid != "" {
			opts = append(opts, "oldoid="+o.OldOid)
		}
		if len(opts) > 0 {
			refSpec = refSpec + "%" + strings.Join(opts, ",")
		}
	}

	if o.RemoteName != "" {
		cmds = append(cmds, o.RemoteName)
	} else {
		cmds = append(cmds, o.RemoteURL)
	}
	cmds = append(cmds, refSpec)

	gitPushCmd.Cmd = cmds[0]
	gitPushCmd.Args = cmds[1:]
	return &gitPushCmd, nil
}

// GetDownloadRef returns reference name of the specific code review.
func (v AGitProtoHelper) GetDownloadRef(cr, patch string) (string, error) {
	_, err := strconv.Atoi(cr)
	if err != nil {
		return "", fmt.Errorf("bad review ID %s: %s", cr, err)
	}
	return fmt.Sprintf("refs/merge-requests/%s/head", cr), nil
}
