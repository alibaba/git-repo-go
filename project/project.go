package project

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
)

// Project inherits manifest's Project and has two related repositories
type Project struct {
	manifest.Project

	RepoRoot         string
	WorkDir          string
	ObjectRepository *Repository
	WorkRepository   *Repository
	manifestURL      string
}

// IsRepoInitialized checks if repository is initialized
func (v *Project) IsRepoInitialized() bool {
	if v.ObjectRepository != nil {
		if !v.ObjectRepository.Exists() {
			return false
		}
	}
	if v.WorkRepository == nil {
		return false
	}
	if !v.WorkRepository.Exists() {
		return false
	}
	return true
}

// GitInit will init project's repositories
func (v *Project) GitInit(manifestURL, referenceGitDir string) error {
	if manifestURL != "" {
		if !strings.HasSuffix(manifestURL, ".git") {
			manifestURL += ".git"
		}
		v.manifestURL = manifestURL
	}
	remoteURL := v.RemoteURL()
	if v.ObjectRepository != nil {
		v.ObjectRepository.Init(v.Remote, remoteURL, referenceGitDir)
	}

	if v.WorkRepository != nil {
		if v.ObjectRepository == nil {
			v.WorkRepository.Init(v.Remote, remoteURL, referenceGitDir)
		} else {
			v.WorkRepository.InitByAttach(v.ObjectRepository)
		}
	}
	return nil
}

func (v *Project) fetchArchive(tarpath string) error {
	cmdArgs := []string{
		"git",
		"archive",
		"-v",
		"-o",
		tarpath,
		"--remote=" + v.RemoteURL(),
		"--prefix=" + v.Path,
		v.Revision,
	}

	return executeCommandIn(v.RepoRoot, cmdArgs)
}

func (v *Project) extractArchive(tarpath string) error {
	cmdArgs := []string{
		"tar",
		"-x",
		"-f",
		tarpath,
	}

	return executeCommandIn(v.RepoRoot, cmdArgs)
}

// Fetch will fetch from remote repository
func (v *Project) Fetch(o *config.FetchOptions) error {
	var err error

	remoteURL := v.RemoteURL()
	if o.Archive && !v.IsMetaProject() {
		if strings.HasPrefix(remoteURL, "http://") ||
			strings.HasPrefix(remoteURL, "https://") {
			return fmt.Errorf("%s: Cannot fetch archives from http/https remotes", v.Name)
		}

		tarpath := strings.ReplaceAll(v.Name, "/", "_")
		tarpath += ".tar"
		err = v.fetchArchive(tarpath)
		if err != nil {
			return fmt.Errorf("fail to fetch tarball %s: %s", tarpath, err)
		}
		err = v.extractArchive(tarpath)
		if err != nil {
			return fmt.Errorf("fail to extract tarball %s: %s", tarpath, err)
		}
		err = os.Remove(filepath.Join(v.RepoRoot, tarpath))
		if err != nil {
			return fmt.Errorf("cannot remove tarball %s: %s", tarpath, err)
		}
		// TODO: CopyAndLinkFiles()
		return nil
	}

	if v.ObjectRepository != nil {
		err = v.ObjectRepository.Fetch(v.Remote)
		if err != nil {
			return err
		}
	}
	if v.WorkRepository != nil {
		err = v.WorkRepository.Fetch(v.Remote)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrepareWorkdir setup workdir layout: .git is gitdir file points to repository
func (v *Project) PrepareWorkdir() error {
	err := os.MkdirAll(v.WorkDir, 0755)
	if err != nil {
		return nil
	}

	gitdir := filepath.Join(v.WorkDir, ".git")
	if _, err = os.Stat(gitdir); err != nil {
		if gitdir != v.WorkRepository.Path {
			relDir, err := filepath.Rel(v.WorkDir, v.WorkRepository.Path)
			if err != nil {
				relDir = v.WorkRepository.Path
			}
			err = ioutil.WriteFile(gitdir,
				[]byte("gitdir: "+relDir+"\n"),
				0644)
			if err != nil {
				return fmt.Errorf("fail to create gitdir for %s: %s", v.Name, err)
			}
		}
	}
	return nil
}

// Checkout will checkout branch
func (v *Project) Checkout(branch, local string) error {
	var (
		err error
		rev string
	)

	err = v.PrepareWorkdir()
	if err != nil {
		return err
	}

	// Run checkout
	if branch == "" {
		if v.Revision != "" {
			branch = v.Revision
		} else {
			branch = "master"
		}
	}
	if strings.HasPrefix(branch, "refs/heads/") {
		branch = strings.TrimPrefix(branch, "refs/heads/")
		rev = fmt.Sprintf("refs/remotes/%s/%s", v.Remote, branch)
	} else {
		rev = branch
	}

	var cmdArgs []string
	if v.Head() != "" {
		cmdArgs = []string{
			"git",
			"rebase",
			rev,
		}
	} else {
		cmdArgs = []string{
			"git",
			"checkout",
		}
		if local != "" {
			cmdArgs = append(cmdArgs, "-b", local)
		}
		cmdArgs = append(cmdArgs, rev)
	}

	err = executeCommandIn(v.WorkDir, cmdArgs)
	if err != nil {
		return fmt.Errorf("fail to checkout %s: %s", v.Name, err)
	}

	return nil
}

// GitRepository returns go-git's repository object for project worktree
func (v *Project) GitRepository() (*git.Repository, error) {
	return git.PlainOpen(v.WorkDir)
}

// GitWorktree returns go-git's worktree oject
func (v *Project) GitWorktree() (*git.Worktree, error) {
	r, err := v.GitRepository()
	if err != nil {
		return nil, err
	}
	return r.Worktree()
}

// Head returns current branch of project's workdir
func (v *Project) Head() string {
	r, err := v.GitRepository()
	if err != nil {
		return ""
	}

	// Not checkout yet
	head, err := r.Head()
	if head == nil {
		return ""
	}
	return head.Name().String()
}

// SetGitRemoteURL sets remote.<remote>.url setting in git config
func (v *Project) SetGitRemoteURL(remote, remoteURL string) error {
	if remote == "" {
		if v.Remote != "" {
			remote = v.Remote
		} else {
			remote = "origin"
		}
	}

	repo := v.ObjectRepository
	if repo == nil {
		repo = v.WorkRepository
	}

	cfg := repo.Config()
	cfg.Set("remote."+remote+".url", remoteURL)
	return repo.SaveConfig(cfg)
}

// GetGitRemoteURL returns remote.<remote>.url setting in git config
func (v *Project) GetGitRemoteURL() string {
	remote := v.Remote
	if remote == "" {
		remote = "origin"
	}

	repo := v.ObjectRepository
	if repo == nil {
		repo = v.WorkRepository
	}

	return repo.Config().Get("remote." + remote + ".url")
}

// RemoteURL returns new remtoe url user provided or from manifest repo url
func (v *Project) RemoteURL() string {
	if v.manifestURL == "" && v.IsMetaProject() {
		v.manifestURL = v.GetGitRemoteURL()
	}
	if v.manifestURL == "" {
		return ""
	}
	if v.IsMetaProject() {
		return v.manifestURL
	}
	if v.GetRemote() == nil {
		log.Errorf("project has no remote: %s", v.Name)
		return ""
	}
	u, err := urlJoin(v.manifestURL, v.GetRemote().Fetch, v.Name)
	if err != nil {
		log.Errorf("fail to join url: %s", err)
		return ""
	}
	return u
}

func urlJoin(manifestURL, fetchURL, name string) (string, error) {
	var (
		u            *url.URL
		err          error
		manglePrefix = false
		mangleColumn = false
	)

	if strings.HasSuffix(manifestURL, "/") {
		manifestURL = strings.TrimRight(manifestURL, "/")
	}
	if strings.HasSuffix(manifestURL, ".git") {
		manifestURL = strings.TrimSuffix(manifestURL, ".git")
	}
	if !strings.Contains(manifestURL, "://") {
		slices := strings.SplitN(manifestURL, ":", 2)
		if len(slices) == 2 {
			manifestURL = strings.Join(slices, "/")
			mangleColumn = true
		}
		manifestURL = "gopher://" + manifestURL
		manglePrefix = true
	}
	u, err = url.Parse(manifestURL)
	if err != nil {
		return "", fmt.Errorf("bad manifest url - %s: %s", manifestURL, err)
	}
	u.Path = filepath.Clean(filepath.Join(u.Path, fetchURL, name))
	joinURL := u.String()

	if manglePrefix {
		joinURL = strings.TrimPrefix(joinURL, "gopher://")
		if mangleColumn {
			slices := strings.SplitN(joinURL, "/", 2)
			if len(slices) == 2 {
				joinURL = strings.Join(slices, ":")
			}
		}
	}
	return joinURL, nil
}

// Config returns git config file parser
func (v *Project) Config() goconfig.GitConfig {
	return v.WorkRepository.Config()
}

// SaveConfig will save config to git config file
func (v *Project) SaveConfig(cfg goconfig.GitConfig) error {
	return v.WorkRepository.SaveConfig(cfg)
}

// NewManifestProject returns a manifest project: a worktree with a seperate repository
func NewManifestProject(repoRoot, manifestURL string) *Project {
	return NewProject(manifest.ManifestsProject, repoRoot, manifestURL)
}

// NewProject returns a project: project worktree with a bared repo and a seperate repository
func NewProject(project *manifest.Project, repoRoot, manifestURL string) *Project {
	var (
		objectRepoPath string
		workRepoPath   string
	)

	p := Project{
		Project:     *project,
		RepoRoot:    repoRoot,
		manifestURL: manifestURL,
	}

	if !p.IsMetaProject() && p.manifestURL == "" {
		log.Panicf("unknown remote url for %s", p.Name)
	}

	if p.IsMetaProject() {
		p.WorkDir = filepath.Join(repoRoot, config.DotRepo, p.Path)
	} else {
		p.WorkDir = filepath.Join(repoRoot, p.Path)
	}

	if !p.IsMetaProject() {
		objectRepoPath = filepath.Join(
			repoRoot,
			config.DotRepo,
			config.ProjectObjects,
			p.Name+".git",
		)
		p.ObjectRepository = &Repository{
			ProjectName: p.Name,
			Path:        objectRepoPath,
			RefSpecs: []string{
				"+refs/heads/*:refs/heads/*",
				"+refs/tags/*:refs/tags/*",
			},
			IsBare: true,
		}
	}

	if p.IsMetaProject() {
		workRepoPath = filepath.Join(
			repoRoot,
			config.DotRepo,
			p.Path+".git",
		)
	} else {
		workRepoPath = filepath.Join(
			repoRoot,
			config.DotRepo,
			config.Projects,
			p.Path+".git",
		)
	}
	p.WorkRepository = &Repository{
		ProjectName: p.Name,
		Path:        workRepoPath,
		RefSpecs: []string{
			"+refs/heads/*:refs/remotes/origin/*",
			"+refs/tags/*:refs/tags/*",
		},
		IsBare: false,
	}

	return &p
}
