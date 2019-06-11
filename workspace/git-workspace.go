package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
)

// GitWorkSpace defines structure for single git workspace
type GitWorkSpace struct {
	RootDir         string
	GitDir          string
	Manifest        *manifest.Manifest
	ManifestProject *project.ManifestProject
	Projects        []*project.Project
	projectByName   map[string][]*project.Project
	projectByPath   map[string]*project.Project
	RemoteMap       map[string]project.RemoteWithError
	httpClient      *http.Client
}

// AdminDir returns .git dir
func (v GitWorkSpace) AdminDir() string {
	return v.GitDir
}

// GetRemoteMap returns RemoteMap
func (v *GitWorkSpace) GetRemoteMap() map[string]project.RemoteWithError {
	return v.RemoteMap
}

// IsSingle is true for git workspace
func (v GitWorkSpace) IsSingle() bool {
	return true
}

// LoadRemotes implements LoadRemotes interface
func (v *GitWorkSpace) LoadRemotes() error {
	if len(v.Projects) != 1 {
		return errors.New("git workspace should contain only one project")
	}
	p := v.Projects[0]
	for _, name := range p.Config().Sections() {
		if strings.HasPrefix(name, "remote.") {
			name = strings.TrimPrefix(name, "remote.")
			remote, err := v.loadRemote(name)
			v.RemoteMap[name] = project.RemoteWithError{Remote: remote, Error: err}
		}
	}
	return nil
}

func (v *GitWorkSpace) loadRemote(name string) (project.Remote, error) {
	if _, ok := v.RemoteMap[name]; ok {
		return v.RemoteMap[name].Remote, v.RemoteMap[name].Error
	}

	p := v.Projects[0]
	repo := p.WorkRepository
	if repo == nil {
		return nil, fmt.Errorf("cannot find repository for project: %s", p.Name)
	}

	remoteURL := repo.GitConfigRemoteURL(name)
	mr := manifest.Remote{
		Name:  name,
		Fetch: remoteURL,
	}

	reviewURL := repo.Config().Get("remote." + name + ".review")
	if reviewURL != "" {
		mr.Review = reviewURL
	} else {
		if remoteURL == "" {
			return nil, fmt.Errorf("upload failed: unknown URL for remote: %s", name)
		}

		gitURL := config.ParseGitURL(remoteURL)
		if gitURL == nil {
			return nil, fmt.Errorf("unsupport git url: %s", remoteURL)
		}
		mr.Review = gitURL.GetReviewURL()
	}

	return loadRemote(&mr)
}

func (v GitWorkSpace) newProject(worktree, gitdir string) (*project.Project, error) {
	name := filepath.Base(worktree)
	s := project.RepoSettings{
		RepoRoot: worktree,
	}

	repo := project.Repository{
		Name:   name,
		Path:   gitdir,
		IsBare: false,
	}

	p := project.Project{
		Project: manifest.Project{
			Name: name,
			Path: ".",
		},

		WorkDir:          worktree,
		ObjectRepository: nil,
		WorkRepository:   &repo,
		Settings:         &s,
	}

	return &p, nil
}

// GetProjects returns all projects
func (v GitWorkSpace) GetProjects(*GetProjectsOptions, ...string) ([]*project.Project, error) {
	return v.Projects, nil
}

// Load sets fields of git work space
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
	v.RemoteMap = make(map[string]project.RemoteWithError)
	v.projectByName[p.Name] = []*project.Project{p}
	v.projectByPath[p.Path] = p
	v.Manifest = nil
	v.ManifestProject = nil

	return nil
}

// NewGitWorkSpace returns workspace interface for single git repository
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
