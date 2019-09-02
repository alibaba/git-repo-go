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

// Package workspace implements the toplevel abstraction of workspace.
package workspace

import (
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/project"
)

// RemoteMap maps name to RemoteWithError
type RemoteMap map[string]project.RemoteWithError

// GetRemote returns remote and error from matching RemoteWithError
func (v RemoteMap) GetRemote(name string) (project.Remote, error) {
	if result, ok := v[name]; ok {
		return result.Remote, result.Error
	}
	return nil, nil
}

// Size is size of map
func (v RemoteMap) Size() int {
	return len(v)
}

// WorkSpace is interface for workspace, implemented with repo workspace or single git workspace.
type WorkSpace interface {
	AdminDir() string
	GetRemoteMap() RemoteMap
	LoadRemotes(bool) error
	IsSingle() bool
	IsMirror() bool
	GetProjects(*GetProjectsOptions, ...string) ([]*project.Project, error)
}

// NewWorkSpace returns workspace instance.
func NewWorkSpace(dir string) (WorkSpace, error) {
	if config.IsSingleMode() {
		return NewGitWorkSpace(dir)
	}
	return NewRepoWorkSpace(dir)
}
