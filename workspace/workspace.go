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
	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/project"
	log "github.com/jiangxin/multi-log"
)

var (
	_ = log.Debug
)

// WorkSpace is interface for workspace, implemented with repo workspace or single git workspace.
type WorkSpace interface {
	AdminDir() string
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
