package project

import (
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
