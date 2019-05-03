package project

import (
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

// GerritRemote is Gerrit remote server
type GerritRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of GerritRemote
func (v *GerritRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of GerritRemote
func (v *GerritRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// GetType returns type of remote
func (v *GerritRemote) GetType() string {
	return config.RemoteTypeGerrit
}
