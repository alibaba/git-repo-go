package workspace

import (
	"errors"
	"net/http"
	"path/filepath"

	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
	log "github.com/jiangxin/multi-log"
)

var (
	_ = log.Debug
)

// GitWorkSpace defines structure for single git workspace.
type GitWorkSpace struct {
	RootDir         string
	GitDir          string
	Manifest        *manifest.Manifest
	ManifestProject *project.ManifestProject
	Projects        []*project.Project
	projectByName   map[string][]*project.Project
	projectByPath   map[string]*project.Project
	httpClient      *http.Client
}

// AdminDir returns .git dir.
func (v GitWorkSpace) AdminDir() string {
	return v.GitDir
}

// IsSingle is true for git workspace.
func (v GitWorkSpace) IsSingle() bool {
	return true
}

// IsMirror is false for git workspace.
func (v GitWorkSpace) IsMirror() bool {
	return false
}

// LoadRemotes implements LoadRemotes interface.
func (v *GitWorkSpace) LoadRemotes(noCache bool) error {
	if len(v.Projects) != 1 {
		return errors.New("git workspace should contain only one project")
	}
	v.Projects[0].LoadRemotes(nil, noCache)
	return nil
}

func (v GitWorkSpace) newProject(worktree, gitdir string) (*project.Project, error) {
	name := filepath.Base(worktree)
	s := project.RepoSettings{
		TopDir: worktree,
	}

	repo := project.Repository{
		Project: manifest.Project{
			Name: name,
			Path: ".",
		},

		DotGit:        gitdir,
		GitDir:        gitdir,
		ObjectsGitDir: "",

		IsBare:   false,
		Settings: &s,
	}

	p := project.Project{
		Repository: repo,

		WorkDir: worktree,
	}

	return &p, nil
}

// GetProjects returns all projects.
func (v GitWorkSpace) GetProjects(*GetProjectsOptions, ...string) ([]*project.Project, error) {
	return v.Projects, nil
}

// Load sets fields of git work space.
func (v *GitWorkSpace) load() error {
	var (
		worktree = v.RootDir
		gitdir   = v.GitDir
	)

	p, err := v.newProject(worktree, gitdir)
	if err != nil {
		return err
	}

	v.Projects = []*project.Project{p}

	v.projectByName = make(map[string][]*project.Project)
	v.projectByPath = make(map[string]*project.Project)
	v.projectByName[p.Name] = []*project.Project{p}
	v.projectByPath[p.Path] = p
	v.Manifest = nil
	v.ManifestProject = nil

	return nil
}

// NewGitWorkSpace returns workspace interface for single git repository.
func NewGitWorkSpace(dir string) (*GitWorkSpace, error) {
	worktree, gitdir, err := path.FindGitWorkSpace(dir)
	if err != nil {
		return nil, err
	}
	return newGitWorkSpace(worktree, gitdir)
}

func newGitWorkSpace(worktree, gitdir string) (*GitWorkSpace, error) {
	var (
		err error
	)

	ws := GitWorkSpace{
		RootDir: worktree,
		GitDir:  gitdir,
	}
	err = ws.load()
	if err != nil {
		return nil, err
	}

	return &ws, nil
}
