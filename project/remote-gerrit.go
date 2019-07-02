package project

import (
	"fmt"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	log "github.com/jiangxin/multi-log"
)

// GerritRemote is Gerrit remote server..
type GerritRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of GerritRemote.
func (v *GerritRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of GerritRemote.
func (v *GerritRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// GetType returns type of remote.
func (v *GerritRemote) GetType() string {
	return config.RemoteTypeGerrit
}

// GetCodeReviewRef returns code review reference: refs/changes/xx/xxxx/<PatchID>.
func (v *GerritRemote) GetCodeReviewRef(reviewID int, patchID int) string {
	if patchID == 0 {
		log.Warn("Patch ID should not be 0, set it to 1")
		patchID = 1
	}
	return fmt.Sprintf("%s%2.2d/%d/%d", config.RefsChanges, reviewID%100, reviewID, patchID)
}

// getReviewURL returns review url.
func (v *GerritRemote) getReviewURL(email string) string {
	var (
		review string
	)

	sshInfo := v.GetSSHInfo()

	if sshInfo == nil || sshInfo.Host == "" || sshInfo.Port == 0 {
		review = v.Review
	} else {
		host := sshInfo.Host
		port := sshInfo.Port
		user := ""
		m := emailUserPattern.FindStringSubmatch(email)
		if len(m) > 1 {
			user = m[1]
		}
		if user == "" {
			user = "git"
		}
		review = fmt.Sprintf("ssh://%s@%s:%d", user, host, port)
	}
	return review
}

// UploadCommands returns upload commands for Gerrit.
func (v *GerritRemote) UploadCommands(o *UploadOptions, branch *ReviewableBranch) ([]string, error) {
	var (
		cmds []string
	)

	p := branch.Project
	url := v.getReviewURL(p.UserEmail())
	if url == "" {
		return nil, fmt.Errorf("review url not configured for '%s'", p.Path)
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += p.Name + ".git"

	gitURL := config.ParseGitURL(url)
	if gitURL == nil {
		return nil, fmt.Errorf("bad review url: %s", url)
	}

	cmds = append(cmds, "git", "push")
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
	uploadType := "for"
	if o.Draft {
		uploadType = "drafts"
	}
	branchName := branch.Branch.Name
	if strings.HasPrefix(branchName, config.RefsHeads) {
		branchName = strings.TrimPrefix(branchName, config.RefsHeads)
	}

	refSpec := fmt.Sprintf("%s:refs/%s/%s",
		config.RefsHeads+branchName,
		uploadType,
		destBranch)
	if o.AutoTopic {
		refSpec = refSpec + "/" + branchName
	}

	opts := []string{}
	for _, u := range o.People[0] {
		opts = append(opts, "r="+u)
	}
	for _, u := range o.People[1] {
		opts = append(opts, "cc="+u)
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

	return cmds, nil
}
