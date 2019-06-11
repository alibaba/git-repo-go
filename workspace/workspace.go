package workspace

import (
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/project"
)

// WorkSpace is interface for repo workspace or single git workspace
type WorkSpace interface {
	AdminDir() string
	GetRemoteMap() map[string]project.RemoteWithError
	LoadRemotes() error
	IsSingle() bool
	GetProjects(*GetProjectsOptions, ...string) ([]*project.Project, error)
}

// NewWorkSpace returns workspace interface
func NewWorkSpace(dir string) (WorkSpace, error) {
	if config.IsSingleMode() {
		return NewGitWorkSpace(dir)
	}
	return NewRepoWorkSpace(dir)
}
