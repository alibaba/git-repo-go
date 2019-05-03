package project

import (
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

// UnknownRemote is Unknown remote server
type UnknownRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of UnknownRemote
func (v *UnknownRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of UnknownRemote
func (v *UnknownRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// GetType returns type of remote
func (v *UnknownRemote) GetType() string {
	return config.RemoteTypeUnknown
}
