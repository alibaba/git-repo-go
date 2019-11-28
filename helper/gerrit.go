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
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/project"
	log "github.com/jiangxin/multi-log"
)

// GerritHelper wraps helper for gerrit server.
type GerritHelper struct {
}

// GetGitPushCommand reads JSON from reader, and format it into proper JSON
// contains git push command.
func (v GerritHelper) GetGitPushCommand(reader io.Reader) ([]byte, error) {
	o := project.UploadOptions{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&o)
	if err != nil {
		return nil, err
	}

	cmds := []string{"git", "push"}

	if o.ReviewURL == "" {
		return nil, fmt.Errorf("review url not configured for '%s'", o.ProjectName)
	}
	if !strings.HasSuffix(o.ReviewURL, "/") {
		o.ReviewURL += "/"
	}
	url := o.ReviewURL + o.ProjectName + ".git"

	gitURL := config.ParseGitURL(url)
	if gitURL == nil {
		return nil, fmt.Errorf("bad review url: %s", url)
	}

	if gitURL.IsSSH() {
		cmds = append(cmds, "--receive-pack=gerrit receive-pack")
	}
	for _, pushOption := range o.PushOptions {
		cmds = append(cmds, "-o", pushOption)
	}
	cmds = append(cmds, url)

	destBranch := o.DestBranch
	if strings.HasPrefix(destBranch, config.RefsHeads) {
		destBranch = strings.TrimPrefix(destBranch, config.RefsHeads)
	}
	if destBranch == "" {
		return nil, fmt.Errorf("empty dest branch for project '%s'", o.ProjectName)
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
	return json.Marshal(&cmd)
}

// GetDownloadRef returns reference name of the specific code review.
func (v GerritHelper) GetDownloadRef(cr, patch string) (string, error) {
	var (
		reviewID int
		patchID  int
		err      error
	)

	reviewID, err = strconv.Atoi(cr)
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
		patchID = 1
	}
	return fmt.Sprintf("%s%2.2d/%d/%d", config.RefsChanges, reviewID%100, reviewID, patchID), nil
}
