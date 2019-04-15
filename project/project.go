package project

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
)

// Project inherits manifest's Project and has two related repositories
type Project struct {
	manifest.Project

	WorkDir          string
	ObjectRepository *Repository
	WorkRepository   *Repository
	Settings         *RepoSettings
}

// FetchOptions is options for git fetch
type FetchOptions struct {
	RepoSettings

	Quiet             bool
	IsNew             bool
	CurrentBranchOnly bool
	CloneBundle       bool
	ForceSync         bool
	NoTags            bool
	OptimizedFetch    bool
	Prune             bool
}

// IsRepoInitialized checks if repository is initialized
func (v Project) IsRepoInitialized() bool {
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

// RepoRoot returns root dir of repo workspace.
func (v Project) RepoRoot() string {
	return v.Settings.RepoRoot
}

// ManifestURL returns manifest URL
func (v Project) ManifestURL() string {
	return v.Settings.ManifestURL
}

// RealPath is pull path of project workdir.
func (v Project) RealPath() string {
	return filepath.Join(v.RepoRoot(), v.Path)
}

// ReferencePath returns path of reference git dir
func (v Project) ReferencePath() string {
	var (
		rdir = ""
		err  error
	)

	if v.Settings.Reference == "" {
		return ""
	}

	if !filepath.IsAbs(v.Settings.Reference) {
		v.Settings.Reference, err = path.Abs(v.Settings.Reference)
		if err != nil {
			log.Errorf("bad reference path '%s': %s", v.Settings.Reference, err)
			v.Settings.Reference = ""
			return ""
		}
	}

	if !v.IsMetaProject() {
		rdir = filepath.Join(v.Settings.Reference, v.Name+".git")
		if path.Exist(rdir) {
			return rdir
		}
		rdir = filepath.Join(v.Settings.Reference,
			config.DotRepo,
			config.Projects,
			v.Path+".git")
		if path.Exist(rdir) {
			return rdir
		}
		return ""
	}

	if v.Settings.ManifestURL != "" {
		u, err := url.Parse(v.Settings.ManifestURL)
		if err == nil {
			dir := u.RequestURI()
			if !strings.HasSuffix(dir, ".git") {
				dir += ".git"
			}
			dirs := strings.Split(dir, "/")
			for i := 1; i < len(dirs); i++ {
				dir = filepath.Join(v.Settings.Reference, filepath.Join(dirs[i:]...))
				if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
					rdir = dir
					break
				}
			}
		}
	}

	if rdir == "" {
		dir := filepath.Join(v.Settings.Reference, config.DotRepo, config.ManifestsDotGit)
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			rdir = dir
		}
	}

	return rdir
}

// Exists indicates whether project exists or not.
func (v Project) Exists() bool {
	if _, err := os.Stat(v.RealPath()); err != nil {
		return false
	}

	if _, err := os.Stat(filepath.Join(v.RealPath(), ".git")); err != nil {
		return false
	}

	return true
}

// GitInit will init project's repositories
func (v *Project) GitInit() error {
	var (
		referenceGitDir string
		remoteURL       string
		err             error
	)

	remoteURL, err = v.GetRemoteURL()
	if err != nil {
		return err
	}

	if v.ObjectRepository != nil {
		v.ObjectRepository.Init("", "", referenceGitDir)
	}

	if v.WorkRepository != nil {
		if v.ObjectRepository == nil {
			v.WorkRepository.Init(v.Remote, remoteURL, referenceGitDir)
		} else {
			v.WorkRepository.InitByLink(v.Remote, remoteURL, v.ObjectRepository)
		}
	}
	return nil
}

func (v *Project) fetchArchive(tarpath string) error {
	u, err := v.GetRemoteURL()
	if err != nil {
		return err
	}
	cmdArgs := []string{
		"git",
		"archive",
		"-v",
		"-o",
		tarpath,
		"--remote=" + u,
		"--prefix=" + v.Path,
		v.Revision,
	}

	return executeCommandIn(v.RepoRoot(), cmdArgs)
}

func (v *Project) extractArchive(tarpath string) error {
	cmdArgs := []string{
		"tar",
		"-x",
		"-f",
		tarpath,
	}

	return executeCommandIn(v.RepoRoot(), cmdArgs)
}

// Fetch will fetch from remote repository
func (v *Project) Fetch(o *FetchOptions) error {
	var err error

	if o == nil {
		o = &FetchOptions{}
	}

	// Initial repository
	v.GitInit()

	remoteURL, err := v.GetRemoteURL()
	if err != nil {
		return err
	}
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
		err = os.Remove(filepath.Join(v.RepoRoot(), tarpath))
		if err != nil {
			return fmt.Errorf("cannot remove tarball %s: %s", tarpath, err)
		}
		// TODO: CopyAndLinkFiles()
		return nil
	}

	if v.WorkRepository != nil {
		err = v.WorkRepository.Fetch(v.Remote, o)
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
	} else if strings.HasPrefix(branch, "refs/") {
		rev = branch
	} else if isHashRevision(branch) {
		rev = branch
	} else {
		rev = fmt.Sprintf("refs/remotes/%s/%s", v.Remote, branch)
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
		if local != "" && local != rev {
			cmdArgs = append(cmdArgs, "-b", local)
		}
		cmdArgs = append(cmdArgs, rev)
	}
	cmdArgs = append(cmdArgs, "--")

	err = executeCommandIn(v.WorkDir, cmdArgs)
	if err != nil {
		return fmt.Errorf("fail to checkout %s, cmd:%s, error: %s",
			v.Name,
			strings.Join(cmdArgs, " "),
			err)
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

// SetManifestURL sets manifestURL and change remote url if is MetaProject
func (v *Project) SetManifestURL(manifestURL string) error {
	if manifestURL != "" && !strings.HasSuffix(manifestURL, ".git") {
		manifestURL += ".git"
	}
	if v.Settings.ManifestURL == manifestURL || manifestURL == "" {
		return nil
	}
	v.Settings.ManifestURL = manifestURL
	if v.IsMetaProject() {
		return v.SetGitRemoteURL(manifestURL)
	}
	return nil
}

// SetGitRemoteURL sets remote.<remote>.url setting in git config
func (v *Project) SetGitRemoteURL(remoteURL string) error {
	remote := v.Remote
	if remote == "" {
		remote = "origin"
	}

	repo := v.WorkRepository
	if repo == nil {
		return fmt.Errorf("project '%s' has no working repository", v.Name)
	}

	cfg := repo.Config()
	cfg.Set("remote."+remote+".url", remoteURL)
	return repo.SaveConfig(cfg)
}

// GitConfigRemoteURL returns remote.<remote>.url setting in git config
func (v *Project) GitConfigRemoteURL() string {
	remote := v.Remote
	if remote == "" {
		remote = "origin"
	}

	repo := v.WorkRepository
	return repo.GitConfigRemoteURL(remote)
}

// GetRemoteURL returns new remtoe url user provided or from manifest repo url
func (v *Project) GetRemoteURL() (string, error) {
	if v.Settings.ManifestURL == "" && v.IsMetaProject() {
		v.Settings.ManifestURL = v.GitConfigRemoteURL()
	}
	if v.Settings.ManifestURL == "" {
		return "", fmt.Errorf("project '%s' has empty manifest url", v.Name)
	}
	if v.IsMetaProject() {
		return v.Settings.ManifestURL, nil
	}
	if v.GetRemote() == nil {
		return "", fmt.Errorf("project '%s' has no remote '%s'", v.Name, v.Remote)
	}

	u, err := urlJoin(v.Settings.ManifestURL, v.GetRemote().Fetch, v.Name+".git")
	if err != nil {
		return "", fmt.Errorf("fail to remote url for '%s': %s", v.Name, err)
	}
	return u, nil
}

// Config returns git config file parser
func (v *Project) Config() goconfig.GitConfig {
	return v.WorkRepository.Config()
}

// SaveConfig will save config to git config file
func (v *Project) SaveConfig(cfg goconfig.GitConfig) error {
	return v.WorkRepository.SaveConfig(cfg)
}

// MatchGroups indecates if project belongs to special groups
func (v Project) MatchGroups(expect string) bool {
	return MatchGroups(expect, v.Groups)
}

// GetSubmoduleProjects returns submodule projects
func (v Project) GetSubmoduleProjects() []*Project {
	// TODO: return submodule projects
	log.Panic("not implement GitSubmodules")
	return nil
}

// NewProject returns a project: project worktree with a bared repo and a seperate repository
func NewProject(project *manifest.Project, s *RepoSettings) *Project {
	var (
		objectRepoPath string
		workRepoPath   string
	)

	// TODO: move to RepoSetting initial function
	if s.ManifestURL != "" && !strings.HasSuffix(s.ManifestURL, ".git") {
		s.ManifestURL += ".git"
	}
	p := Project{
		Project:  *project,
		Settings: s,
	}

	if !p.IsMetaProject() && s.ManifestURL == "" {
		log.Panicf("unknown remote url for %s", p.Name)
	}

	if p.IsMetaProject() {
		p.WorkDir = filepath.Join(p.RepoRoot(), config.DotRepo, p.Path)
	} else {
		p.WorkDir = filepath.Join(p.RepoRoot(), p.Path)
	}

	if !p.IsMetaProject() && cap.Symlink() {
		objectRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			config.ProjectObjects,
			p.Name+".git",
		)
		p.ObjectRepository = &Repository{
			Name:     p.Name,
			RelPath:  p.Path,
			Path:     objectRepoPath,
			IsBare:   true,
			Settings: s,
		}
	}

	if p.IsMetaProject() {
		workRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			p.Path+".git",
		)
	} else {
		workRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			config.Projects,
			p.Path+".git",
		)
	}

	p.WorkRepository = &Repository{
		Name:      p.Name,
		RelPath:   p.Path,
		Path:      workRepoPath,
		Remote:    p.Remote,
		Revision:  p.Revision,
		IsBare:    false,
		Reference: p.ReferencePath(),
		Settings:  s,
	}

	remoteURL, err := p.GetRemoteURL()
	if err != nil {
		log.Panicf("fail to get remote url for '%s': %s", p.Name, err)
	}
	p.WorkRepository.RemoteURL = remoteURL

	return &p
}

func isHashRevision(rev string) bool {
	re := regexp.MustCompile(`^[0-9][a-f]{7,}$`)
	return re.MatchString(rev)
}

// Join two group of projects, ignore duplicated projects
func Join(group1, group2 []*Project) []*Project {
	projectMap := make(map[string]bool)
	result := make([]*Project, len(group1)+len(group2))

	for _, p := range group1 {
		if _, ok := projectMap[p.Path]; !ok {
			projectMap[p.Path] = true
			result = append(result, p)
		}
	}
	for _, p := range group2 {
		if _, ok := projectMap[p.Path]; !ok {
			projectMap[p.Path] = true
			result = append(result, p)
		}
	}
	return result
}

// GroupByName returns a map using project name as index to group projects
func GroupByName(projects []*Project) map[string][]*Project {
	result := make(map[string][]*Project)
	for _, p := range projects {
		if _, ok := result[p.Name]; !ok {
			result[p.Name] = []*Project{p}
		} else {
			result[p.Name] = append(result[p.Name], p)
		}
	}
	return result
}

// PathEntry is used to group projects by path
type PathEntry struct {
	Path    string
	Project *Project
	Entries []*PathEntry
}

// GroupByPath returns a map using project path as index to group projects
func GroupByPath(projects []*Project) *PathEntry {
	pMap := make(map[string]*Project)
	paths := []string{}
	for _, p := range projects {
		pMap[p.Path] = p
		paths = append(paths, p.Path)
	}
	sort.Strings(paths)

	rootEntry := &PathEntry{Path: "/"}
	appendEntries(rootEntry, paths, pMap)
	return rootEntry
}

func appendEntries(entry *PathEntry, paths []string, pMap map[string]*Project) {
	var (
		oldEntry, newEntry *PathEntry
		i, j               int
	)

	for i = 0; i < len(paths); i++ {
		for strings.HasSuffix(paths[i], "/") {
			paths[i] = strings.TrimSuffix(paths[i], "/")
		}
		newEntry = &PathEntry{
			Path:    paths[i],
			Project: pMap[paths[i]],
		}
		if i > 0 && strings.HasPrefix(paths[i], paths[i-1]+"/") {
			for j = i + 1; j < len(paths); j++ {
				if !strings.HasPrefix(paths[j], paths[i-1]+"/") {
					break

				}
			}
			appendEntries(oldEntry, paths[i:j], pMap)
			i = j - 1
			continue
		}
		oldEntry = newEntry
		entry.Entries = append(entry.Entries, newEntry)
	}
}
