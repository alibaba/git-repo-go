package project

import (
	"fmt"

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

// GetCodeReviewRef returns code review reference: refs/merge-requests/<ID>/head
func (v *UnknownRemote) GetCodeReviewRef(reviewID int, patchID int) string {
	return ""
}

// UploadCommands returns upload commands
func (v *UnknownRemote) UploadCommands(o *UploadOptions, branch *ReviewableBranch) ([]string, error) {
	return nil, fmt.Errorf("unknown remote for upload")
}
