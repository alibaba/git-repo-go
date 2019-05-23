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
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Project inherits manifest's Project and has two related repositories
type Project struct {
	manifest.Project

	WorkDir          string
	ObjectRepository *Repository
	WorkRepository   *Repository
	Settings         *RepoSettings
	Remote           Remote
}

// ConfigWithDefault checks git config from both project and manifest project
type ConfigWithDefault struct {
	Project *Project
}

// HasKey checks whether key is set in both project and manifest project
func (v ConfigWithDefault) HasKey(key string) bool {
	value := v.Project.Config().HasKey(key)
	if value || v.Project.IsMetaProject() {
		return value
	}
	if v.Project.Settings != nil && v.Project.Settings.Config != nil {
		return v.Project.Settings.Config.HasKey(key)
	}
	return false
}

// Get returns config of both project and manifest project
func (v ConfigWithDefault) Get(key string) string {
	value := v.Project.Config().Get(key)
	if value != "" || v.Project.IsMetaProject() {
		return value
	}
	if v.Project.Settings != nil && v.Project.Settings.Config != nil {
		return v.Project.Settings.Config.Get(key)
	}
	return ""
}

// GetBool returns boolean config of both project and manifest project
func (v ConfigWithDefault) GetBool(key string, defaultVal bool) bool {
	value := v.Get(key)
	switch strings.ToLower(value) {
	case "yes", "true", "on":
		return true
	case "no", "false", "off":
		return false
	}
	return defaultVal
}

// RepoRoot returns root dir of repo workspace.
func (v Project) RepoRoot() string {
	return v.Settings.RepoRoot
}

// ManifestURL returns manifest URL
func (v Project) ManifestURL() string {
	return v.Settings.ManifestURL
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
	if _, err := os.Stat(v.WorkDir); err != nil {
		return false
	}

	if v.Settings.Mirror {
		return true
	}

	if _, err := os.Stat(filepath.Join(v.WorkDir, ".git")); err != nil {
		return false
	}

	return true
}

// PrepareWorkdir setup workdir layout: .git is gitdir file points to repository
func (v *Project) PrepareWorkdir() error {
	err := os.MkdirAll(v.WorkDir, 0755)
	if err != nil {
		return nil
	}

	gitdir := filepath.Join(v.WorkDir, ".git")
	if _, err = os.Stat(gitdir); err != nil {
		// Remove index file for fresh checkout
		idxfile := filepath.Join(v.WorkRepository.Path, "index")
		err = os.Remove(idxfile)

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
	return nil
}

// CleanPublishedCache removes obsolete refs/published/ references.
func (v Project) CleanPublishedCache() error {
	var err error

	raw := v.WorkRepository.Raw()
	if raw == nil {
		return nil
	}
	pubMap := make(map[string]string)
	headsMap := make(map[string]string)
	refs, _ := raw.References()
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			if strings.HasPrefix(string(ref.Name()), config.RefsPub) {
				pubMap[string(ref.Name())] = ref.Hash().String()
			} else if strings.HasPrefix(string(ref.Name()), config.RefsHeads) {
				headsMap[string(ref.Name())] = ref.Hash().String()
			}
			fmt.Println(ref)
		}
		return nil
	})

	for name := range pubMap {
		branch := config.RefsHeads + strings.TrimPrefix(name, config.RefsPub)
		if _, ok := headsMap[branch]; !ok {
			log.Infof("will delete obsolete ref: %s", name)
			err = raw.Storer.RemoveReference(plumbing.ReferenceName(name))
			if err != nil {
				log.Errorf("fail to remove reference '%s'", name)
			} else {
				log.Infof("removed reference '%s'", name)
			}
		}
	}

	return nil
}

// IsRebaseInProgress checks whether is in middle of a rebase.
func (v Project) IsRebaseInProgress() bool {
	return v.WorkRepository.IsRebaseInProgress()
}

// RevisionIsValid checks if revision is valid
func (v Project) RevisionIsValid(revision string) bool {
	return v.WorkRepository.RevisionIsValid(revision)
}

// Revlist works like rev-list
func (v Project) Revlist(args ...string) ([]string, error) {
	return v.WorkRepository.Revlist(args...)
}

// PublishedReference forms published reference for specific branch.
func (v Project) PublishedReference(branch string) string {
	pub := config.RefsPub + branch

	if v.RevisionIsValid(pub) {
		return pub
	}
	return ""
}

// PublishedRevision resolves published reference to revision id.
func (v Project) PublishedRevision(branch string) string {
	raw := v.WorkRepository.Raw()
	pub := config.RefsPub + branch

	if raw == nil {
		return ""
	}

	revid, err := raw.ResolveRevision(plumbing.Revision(pub))
	if err == nil {
		return revid.String()
	}
	return ""
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
	remote := v.RemoteName
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
	remote := v.RemoteName
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
	if v.ManifestRemote == nil {
		return "", fmt.Errorf("project '%s' has no remote '%s'", v.Name, v.RemoteName)
	}

	u, err := urlJoin(v.Settings.ManifestURL, v.ManifestRemote.Fetch, v.Name+".git")
	if err != nil {
		return "", fmt.Errorf("fail to remote url for '%s': %s", v.Name, err)
	}
	return u, nil
}

// Config returns git config file parser
func (v *Project) Config() goconfig.GitConfig {
	return v.WorkRepository.Config()
}

// ConfigWithDefault returns git config file parser
func (v *Project) ConfigWithDefault() ConfigWithDefault {
	return ConfigWithDefault{Project: v}
}

// ManifestConfig returns git config of manifest project
func (v *Project) ManifestConfig() goconfig.GitConfig {
	return v.Settings.Config
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

// UserEmail returns user identity
func (v Project) UserEmail() string {
	username := os.Getenv("GIT_COMMITTER_NAME")
	useremail := os.Getenv("GIT_COMMITTER_EMAIL")
	if username == "" {
		username = v.ConfigWithDefault().Get("user.name")
	}
	if useremail == "" {
		useremail = v.ConfigWithDefault().Get("user.email")
	}
	if username != "" && useremail != "" {
		if strings.Contains(username, " ") {
			return fmt.Sprintf("\"%s\" <%s>", username, useremail)
		}
		return fmt.Sprintf("%s <%s>", username, useremail)
	}
	return ""
}

// NewProject returns a project: project worktree with a bared repo and a seperate repository
func NewProject(project *manifest.Project, s *RepoSettings) *Project {
	var (
		objectRepoPath string
		workRepoPath   string
	)

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

	if !p.IsMetaProject() && cap.CanSymlink() {
		objectRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			config.ProjectObjects,
			p.Name+".git",
		)
		p.ObjectRepository = &Repository{
			Name:     p.Name,
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
		Name:       p.Name,
		Path:       workRepoPath,
		RemoteName: p.RemoteName,
		Revision:   p.Revision,
		IsBare:     false,
		Reference:  p.ReferencePath(),
		Settings:   s,
	}

	remoteURL, err := p.GetRemoteURL()
	if err != nil {
		log.Panicf("fail to get remote url for '%s': %s", p.Name, err)
	}
	p.WorkRepository.RemoteURL = remoteURL

	return &p
}

// NewMirrorProject returns a mirror project
func NewMirrorProject(project *manifest.Project, s *RepoSettings) *Project {
	var (
		repoPath string
	)

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

	repoPath = filepath.Join(
		p.RepoRoot(),
		p.Name+".git",
	)

	p.WorkDir = repoPath

	p.WorkRepository = &Repository{
		Name:       p.Name,
		Path:       repoPath,
		RemoteName: p.RemoteName,
		Revision:   p.Revision,
		IsBare:     true,
		Reference:  p.ReferencePath(),
		Settings:   s,
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

// IndexByName returns a map using project name as index to group projects
func IndexByName(projects []*Project) map[string][]*Project {
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

// IndexByPath returns a map using project path as index to group projects
func IndexByPath(projects []*Project) map[string]*Project {
	result := make(map[string]*Project)
	for _, p := range projects {
		result[p.Path] = p
	}
	return result
}

// Tree is used to group projects by path
type Tree struct {
	Path    string
	Project *Project
	Trees   []*Tree
}

// ProjectsTree returns a map using project path as index to group projects
func ProjectsTree(projects []*Project) *Tree {
	pMap := make(map[string]*Project)
	paths := []string{}
	for _, p := range projects {
		pMap[p.Path] = p
		paths = append(paths, p.Path)
	}
	sort.Strings(paths)

	root := &Tree{Path: "/"}
	treeAppend(root, paths, pMap)
	return root
}

func treeAppend(tree *Tree, paths []string, pMap map[string]*Project) {
	var (
		oldEntry, newEntry *Tree
		i, j               int
	)

	for i = 0; i < len(paths); i++ {
		for strings.HasSuffix(paths[i], "/") {
			paths[i] = strings.TrimSuffix(paths[i], "/")
		}
		newEntry = &Tree{
			Path:    paths[i],
			Project: pMap[paths[i]],
		}
		if i > 0 && strings.HasPrefix(paths[i], paths[i-1]+"/") {
			for j = i + 1; j < len(paths); j++ {
				if !strings.HasPrefix(paths[j], paths[i-1]+"/") {
					break

				}
			}
			treeAppend(oldEntry, paths[i:j], pMap)
			i = j - 1
			continue
		}
		oldEntry = newEntry
		tree.Trees = append(tree.Trees, newEntry)
	}
}

// DetachHead makes detached HEAD
func (v Project) DetachHead() error {
	cmdArgs := []string{
		GIT,
		"checkout",
		"HEAD^0",
		"--",
	}
	return executeCommandIn(v.WorkDir, cmdArgs)
}
