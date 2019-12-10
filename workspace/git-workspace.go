package workspace

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/helper"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
	log "github.com/jiangxin/multi-log"
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
	RemoteMap       map[string]project.Remote
	httpClient      *http.Client
}

// AdminDir returns .git dir.
func (v GitWorkSpace) AdminDir() string {
	return v.GitDir
}

// GetRemoteMap returns RemoteMap.
func (v *GitWorkSpace) GetRemoteMap() RemoteMap {
	return v.RemoteMap
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
	var (
		query *helper.SSHInfoQuery
	)
	if len(v.Projects) != 1 {
		return errors.New("git workspace should contain only one project")
	}
	p := v.Projects[0]
	cfg := p.Config()
	for _, name := range cfg.Sections() {
		if !strings.HasPrefix(name, "remote.") {
			continue
		}

		name = strings.TrimPrefix(name, "remote.")
		remoteURL := p.GitConfigRemoteURL(name)
		if remoteURL == "" {
			log.Warnf("no URL defined for remote: %s", name)
			continue
		}
		log.Debugf("URL of remote %s: %s", name, remoteURL)
		mr := manifest.Remote{
			Name:  name,
			Fetch: remoteURL,
		}
		reviewURL := cfg.Get("remote." + name + ".review")
		if reviewURL != "" {
			mr.Review = reviewURL
		} else {
			gitURL := config.ParseGitURL(remoteURL)
			if gitURL == nil {
				log.Debugf("fail to parse remote: %s, URL: %s", name, remoteURL)
				continue
			}
			reviewURL = gitURL.GetReviewURL()
			if reviewURL == "" {
				log.Debugf("cannot get review URL from remote: %s, URL: %s", name, remoteURL)
				continue
			}
			mr.Review = reviewURL
		}
		log.Debugf("review of remote %s is: %s", name, reviewURL)

		query = helper.NewSSHInfoQuery(p.SSHInfoCacheFile())
		sshInfo, err := query.GetSSHInfo(mr.Review, !noCache)
		if err != nil {
			return err
		}
		protoHelper := helper.NewProtoHelper(sshInfo)
		remote := project.NewRemote(&mr, protoHelper)
		log.Debugf("loaded remote: %#v, error: %s", remote, err)
		v.RemoteMap[mr.Name] = *remote
	}

	if len(v.RemoteMap) == 1 {
		for name := range v.RemoteMap {
			v.Projects[0].Remote = v.RemoteMap[name]
		}
	}

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
	v.RemoteMap = make(map[string]project.Remote)
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
