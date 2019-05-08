package project

import (
	"encoding/base64"
	"fmt"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

// AGitRemote is AGit remote server
type AGitRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of AGitRemote
func (v *AGitRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of AGitRemote
func (v *AGitRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// GetType returns type of remote
func (v *AGitRemote) GetType() string {
	return config.RemoteTypeAGit
}

// getReviewURL returns review url
func (v *AGitRemote) getReviewURL(email string) string {
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

// UploadCommands returns upload commands for AGit
func (v *AGitRemote) UploadCommands(o *UploadOptions, branch *ReviewableBranch) ([]string, error) {
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

	cmds = append(cmds, "git", "push")
	if strings.HasPrefix(url, "ssh://") {
		cmds = append(cmds, "--receive-pack=agit-receive-pack")
	}
	for _, pushOption := range o.PushOptions {
		cmds = append(cmds, "-o", pushOption)
	}

	if o.Title != "" {
		cmds = append(cmds, "-o",
			fmt.Sprintf("title={base64}%s",
				base64.StdEncoding.EncodeToString([]byte(o.Title))))
	}
	if o.Description != "" {
		cmds = append(cmds, "-o",
			fmt.Sprintf("description={base64}%s",
				base64.StdEncoding.EncodeToString([]byte(o.Description))))
	}
	if o.Issue != "" {
		cmds = append(cmds, "-o", fmt.Sprintf("issue=%s", o.Issue))
	}

	if len(o.People[0]) > 0 {
		reviewers := strings.Join(o.People[0], ",")
		cmds = append(cmds, "-o", "reviewers="+reviewers)
	}
	if len(o.People[1]) > 0 {
		cc := strings.Join(o.People[1], ",")
		cmds = append(cmds, "-o", "cc="+cc)
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

	refSpec := fmt.Sprintf("%s:refs/%s/%s/%s",
		config.RefsHeads+branchName,
		uploadType,
		destBranch,
		branchName)

	cmds = append(cmds, refSpec)

	return cmds, nil
}
