package project

import (
	"code.alibaba-inc.com/force/git-repo/helper"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

// Remote wraps manifest remote.
type Remote struct {
	manifest.Remote
	helper.ProtoHelper

	initialized bool
}

func (v *Remote) Initialized() bool {
	return v.initialized
}

// NewRemote return new remote object.
func NewRemote(r *manifest.Remote, h helper.ProtoHelper) *Remote {
	if h == nil {
		h = &helper.DefaultProtoHelper{}
	}
	remote := Remote{
		Remote:      *r,
		ProtoHelper: h,

		initialized: true,
	}

	return &remote
}
