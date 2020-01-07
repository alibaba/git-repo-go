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

	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/manifest"
	"github.com/aliyun/git-repo-go/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Project inherits manifest's Project and has two related repositories.
type Project struct {
	Repository

	WorkDir string
}

// ConfigWithDefault checks git config from both project and manifest project.
type ConfigWithDefault struct {
	Project *Project
}

// HasKey checks whether key is set in both project and manifest project.
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

// Get returns config of both project and manifest project.
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

// GetBool returns boolean config of both project and manifest project.
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

// TopDir returns root dir of repo workspace.
func (v Project) TopDir() string {
	return v.Settings.TopDir
}

// ManifestURL returns manifest URL.
func (v Project) ManifestURL() string {
	return v.Settings.ManifestURL
}

func referencePath(mp *manifest.Project, s *RepoSettings) string {
	var (
		rdir = ""
		err  error
	)

	if s.Reference == "" {
		return ""
	}

	if !filepath.IsAbs(s.Reference) {
		s.Reference, err = path.Abs(s.Reference)
		if err != nil {
			log.Errorf("bad reference path '%s': %s", s.Reference, err)
			s.Reference = ""
			return ""
		}
	}

	if !mp.IsMetaProject() {
		rdir = filepath.Join(s.Reference, mp.Name+".git")
		if path.Exist(rdir) {
			return rdir
		}
		rdir = filepath.Join(s.Reference,
			config.DotRepo,
			config.Projects,
			mp.Path+".git")
		if path.Exist(rdir) {
			return rdir
		}
		return ""
	}

	if s.ManifestURL != "" {
		u, err := url.Parse(s.ManifestURL)
		if err == nil {
			dir := u.RequestURI()
			if !strings.HasSuffix(dir, ".git") {
				dir += ".git"
			}
			dirs := strings.Split(dir, "/")
			for i := 1; i < len(dirs); i++ {
				dir = filepath.Join(s.Reference, filepath.Join(dirs[i:]...))
				if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
					rdir = dir
					break
				}
			}
		}
	}

	if rdir == "" {
		dir := filepath.Join(s.Reference, config.DotRepo, config.ManifestsDotGit)
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			rdir = dir
		}
	}

	return rdir
}

// IsMirror indicates project is a mirror repository, no checkout
func (v Project) IsMirror() bool {
	return v.Settings.Mirror
}

// Exists indicates whether project exists or not.
func (v Project) Exists() bool {
	if path.IsDir(v.GitDir) && path.IsDir(v.ObjectsGitDir) {
		return true
	}
	return false
}

// PrepareWorkdir setup workdir layout: .git is gitdir file points to repository.
func (v *Project) PrepareWorkdir() error {
	err := os.MkdirAll(v.WorkDir, 0755)
	if err != nil {
		return nil
	}

	gitdir := filepath.Join(v.WorkDir, ".git")
	if _, err = os.Stat(gitdir); err != nil {
		// Remove index file for fresh checkout
		idxfile := filepath.Join(v.GitDir, "index")
		err = os.Remove(idxfile)

		relDir, err := filepath.Rel(v.WorkDir, v.GitDir)
		if err != nil {
			relDir = v.GitDir
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

	raw := v.Raw()
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
		}
		return nil
	})

	for name := range pubMap {
		branch := config.RefsHeads + strings.TrimPrefix(name, config.RefsPub)
		if _, ok := headsMap[branch]; !ok {
			log.Infof("%swill delete obsolete ref: %s", v.Prompt(), name)
			err = raw.Storer.RemoveReference(plumbing.ReferenceName(name))
			if err != nil {
				log.Errorf("%sfail to remove reference '%s'", v.Prompt(), name)
			} else {
				log.Infof("%sremoved reference '%s'", v.Prompt(), name)
			}
		}
	}

	return nil
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
	raw := v.Raw()
	if strings.HasPrefix(branch, config.RefsHeads) {
		branch = strings.TrimPrefix(branch, config.RefsHeads)
	}
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

// GitRepository returns go-git's repository object for project worktree.
func (v *Project) GitRepository() (*git.Repository, error) {
	return git.PlainOpen(v.WorkDir)
}

// GitWorktree returns go-git's worktree oject.
func (v *Project) GitWorktree() (*git.Worktree, error) {
	r, err := v.GitRepository()
	if err != nil {
		return nil, err
	}
	return r.Worktree()
}

// HeadBranch returns current branch (name and oid) of project's workdir.
func (v *Project) HeadBranch() Branch {
	r, err := v.GitRepository()
	if err != nil {
		return Branch{}
	}

	// Not checkout yet
	head, _ := r.Head()
	if head == nil {
		return Branch{}
	}

	headName := head.Name().String()
	if headName == "HEAD" {
		return Branch{
			Name: "",
			Hash: head.Hash().String(),
		}
	}
	return Branch{
		Name: head.Name().String(),
		Hash: head.Hash().String(),
	}
}

// SetManifestURL sets manifestURL and change remote url if is MetaProject.
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

// SetGitRemoteURL sets remote.<remote>.url setting in git config.
func (v *Project) SetGitRemoteURL(remoteURL string) error {
	remote := v.RemoteName
	if remote == "" {
		remote = "origin"
	}

	if !v.Repository.Exists() {
		return fmt.Errorf("project '%s' has no working repository", v.Name)
	}

	cfg := v.Config()
	cfg.Set("remote."+remote+".url", remoteURL)
	return v.SaveConfig(cfg)
}

// DisableDefaultPush sets git config push.default to nothing.
func (v *Project) DisableDefaultPush() error {
	cfg := v.Config()
	if cfg.Get("push.default") == "" {
		log.Debugf("%sdisable default push by setting push.default to noting", v.Prompt())
		cfg.Set("push.default", "nothing")
		return v.SaveConfig(cfg)
	}
	return nil
}

func (v *Project) gitConfigRemoteURL() string {
	remote := v.RemoteName
	if remote == "" {
		remote = "origin"
	}

	return v.GitConfigRemoteURL(remote)
}

// GetRemoteURL returns new remtoe url user provided or from manifest repo url
func (v *Project) GetRemoteURL() (string, error) {
	if v.Settings.ManifestURL == "" && v.IsMetaProject() {
		v.Settings.ManifestURL = v.gitConfigRemoteURL()
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

// ConfigWithDefault returns git config file parser.
func (v *Project) ConfigWithDefault() ConfigWithDefault {
	return ConfigWithDefault{Project: v}
}

// ManifestConfig returns git config of manifest project.
func (v *Project) ManifestConfig() goconfig.GitConfig {
	return v.Settings.Config
}

// MatchGroups indecates if project belongs to special groups.
func (v Project) MatchGroups(expect string) bool {
	return MatchGroups(expect, v.Groups)
}

// GetSubmoduleProjects returns submodule projects.
func (v Project) GetSubmoduleProjects() []*Project {
	// TODO: return submodule projects
	log.Fatal("not implement GitSubmodules")
	return nil
}

// UserEmail returns user identity.
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

// NewProject returns a project: project worktree with a bared repo and a seperate repository.
func NewProject(mp *manifest.Project, s *RepoSettings) *Project {
	var (
		workDir       string
		dotGit        string
		gitDir        string
		objectsGitDir string
	)

	if !mp.IsMetaProject() && s.ManifestURL == "" {
		log.Panicf("unknown remote url for %s", mp.Name)
	}

	if s.ManifestURL != "" && !strings.HasSuffix(s.ManifestURL, ".git") {
		s.ManifestURL += ".git"
	}

	if mp.IsMetaProject() {
		workDir = filepath.Join(s.TopDir, config.DotRepo, mp.Path)
		gitDir = filepath.Join(
			s.TopDir,
			config.DotRepo,
			mp.Path+".git",
		)
		objectsGitDir = ""
	} else {
		workDir = filepath.Join(s.TopDir, mp.Path)
		gitDir = filepath.Join(
			s.TopDir,
			config.DotRepo,
			config.Projects,
			mp.Path+".git",
		)
		objectsGitDir = filepath.Join(
			s.TopDir,
			config.DotRepo,
			config.ProjectObjects,
			mp.Name+".git",
		)
	}
	dotGit = filepath.Join(workDir, ".git")

	repo := Repository{
		Project: *mp,

		DotGit:        dotGit,
		GitDir:        gitDir,
		ObjectsGitDir: objectsGitDir,

		IsBare:    false,
		Settings:  s,
		Reference: referencePath(mp, s),
		Remotes:   NewRemoteMap(),
	}

	p := Project{
		Repository: repo,
		WorkDir:    workDir,
	}

	remoteURL, err := p.GetRemoteURL()
	if err != nil {
		log.Panicf("fail to get remote url for '%s': %s", p.Name, err)
	}
	p.Repository.RemoteURL = remoteURL

	return &p
}

// NewMirrorProject returns a mirror project.
func NewMirrorProject(mp *manifest.Project, s *RepoSettings) *Project {
	var (
		gitDir string
	)

	if s.ManifestURL != "" && !strings.HasSuffix(s.ManifestURL, ".git") {
		s.ManifestURL += ".git"
	}

	gitDir = filepath.Join(
		s.TopDir,
		mp.Name+".git",
	)

	repo := Repository{
		Project: *mp,

		DotGit:        "",
		GitDir:        gitDir,
		ObjectsGitDir: gitDir,

		IsBare:    true,
		Settings:  s,
		Reference: referencePath(mp, s),
	}

	p := Project{
		Repository: repo,
		WorkDir:    "",
	}

	if !mp.IsMetaProject() && s.ManifestURL == "" {
		log.Panicf("unknown remote url for %s", mp.Name)
	}

	remoteURL, err := p.GetRemoteURL()
	if err != nil {
		log.Panicf("fail to get remote url for '%s': %s", p.Name, err)
	}
	p.Repository.RemoteURL = remoteURL

	return &p
}

func isHashRevision(rev string) bool {
	re := regexp.MustCompile(`^[0-9][a-f]{7,}$`)
	return re.MatchString(rev)
}

// Join two group of projects, ignore duplicated projects.
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

// IndexByName returns a map using project name as index to group projects.
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

// IndexByPath returns a map using project path as index to group projects.
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

// ProjectsTree returns a map using project path as index to group projects.
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

// DetachHead makes detached HEAD.
func (v Project) DetachHead() error {
	cmdArgs := []string{
		GIT,
		"checkout",
		"-q",
		"HEAD^0",
		"--",
	}
	return executeCommandIn(v.WorkDir, cmdArgs)
}
