package workspace

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
	"github.com/jiangxin/multi-log"
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
	RemoteMap       map[string]project.Remote
	httpClient      *http.Client
}

// LoadRemotes implements LoadRemotes interface, do nothing
func (v *GitWorkSpace) LoadRemotes() error {
	return nil
}

func (v GitWorkSpace) newProject(worktree, gitdir string) (*project.Project, error) {
	var (
		remoteName     = ""
		remoteRevision = ""
		remoteURL      = ""
		err            error
	)

	name := filepath.Base(worktree)
	s := project.RepoSettings{
		RepoRoot: worktree,
	}

	repo := project.Repository{
		Name:   name,
		Path:   gitdir,
		IsBare: false,
	}

	head := repo.GetHead()
	if project.IsHead(head) {
		head = strings.TrimPrefix(head, config.RefsHeads)
		remoteName = repo.TrackRemote(head)
		remoteRevision = repo.TrackBranch(head)
		if remoteName == "" || remoteRevision == "" {
			return nil, fmt.Errorf("upload failed: cannot find tracking branch\n\n" +
				"Please run command \"git branch -u <upstream>\" to track a remote branch. E.g.:\n\n" +
				"    git branch -u origin/master")
		}
		remoteURL = repo.GitConfigRemoteURL(remoteName)
		if remoteURL == "" {
			return nil, fmt.Errorf("upload failed: unknown URL for remote: %s", remoteName)
		}
		repo.RemoteName = remoteName
		repo.Revision = remoteRevision
		repo.RemoteURL = remoteURL
	} else {
		log.Debugf("detached at %s", head)
		return nil, fmt.Errorf("upload failed: not in a branch\n\n" +
			"Please run command \"git checkout -b <branch>\" to create a new branch.")
	}

	p := project.Project{
		Project: manifest.Project{
			Name:       name,
			Path:       ".",
			RemoteName: remoteName,
			Revision:   remoteRevision,
		},

		WorkDir:          worktree,
		ObjectRepository: nil,
		WorkRepository:   &repo,
		Settings:         &s,
	}

	reviewURL := p.Config().Get("remote." + remoteName + ".review")
	if reviewURL == "" {
		reviewURL, err = getReviewURL(remoteURL)
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("Review URL: %s", reviewURL)

	mr := manifest.Remote{
		Name:     remoteName,
		Fetch:    remoteURL,
		Revision: remoteRevision,
		Review:   reviewURL,
	}

	remote, err := loadRemote(&mr)
	if err != nil {
		return nil, err
	}

	p.Remote = remote
	return &p, nil
}

// GetProjects returns all projects
func (v GitWorkSpace) GetProjects(*GetProjectsOptions, ...string) ([]*project.Project, error) {
	return v.Projects, nil
}

func getReviewURL(address string) (string, error) {
	var (
		unknownProtoForReview = regexp.MustCompile(`^(.+)://.+`)

		match []string
	)

	match = config.GitHTTPProtocolPattern.FindStringSubmatch(address)
	if len(match) > 0 {
		return match[1] + match[3], nil
	}
	match = config.GitSSHProtocolPattern.FindStringSubmatch(address)
	if len(match) > 0 {
		return match[2], nil
	}
	match = config.GitRsyncProtocolPattern.FindStringSubmatch(address)
	if len(match) > 0 {
		return match[2], nil
	}
	match = unknownProtoForReview.FindStringSubmatch(address)
	if len(match) > 0 {
		return "", fmt.Errorf("cannot find review URL for protocol: '%s'", match[1])
	}
	return "", fmt.Errorf("cannot find review URL from '%s'", address)
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
	v.RemoteMap = make(map[string]project.Remote)
	v.projectByName[p.Name] = []*project.Project{p}
	v.projectByPath[p.Path] = p
	if p.Remote != nil {
		r := p.Remote.GetRemote()
		if r != nil {
			v.RemoteMap[r.Name] = p.Remote
		}
	}
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
