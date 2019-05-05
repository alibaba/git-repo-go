package project

import (
	"fmt"
	"regexp"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

var (
	emailUserPattern = regexp.MustCompile(`^.* <([^\s]+)@[^\s]+>$`)
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
