package project

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Macros for project package
const (
	GIT = "git"
)

// Default settings
var (
	DefaultRefSpecs = []string{
		"+refs/heads/*:refs/heads/*",
		"+refs/tags/*:refs/tags/*",
	}
)

// Repository has repository related operations
type Repository struct {
	Name      string // Project name
	RelPath   string // Project relative path
	Path      string // Repository real path
	RefSpecs  []string
	IsBare    bool
	RemoteURL string
	Reference string
	Remote    string // Project Remote field from manifest xml
	Revision  string // Projeect Revision from manifest xml
	Settings  *RepoSettings
	raw       *git.Repository
}

// Exists checks repository layout
func (v *Repository) Exists() bool {
	return path.IsGitDir(v.Path)
}

func (v *Repository) setRemote(remoteName, remoteURL string) error {
	var err error

	if remoteURL != "" {
		v.RemoteURL = remoteURL
	}
	cfg := v.Config()
	changed := false
	if !v.IsBare {
		cfg.Unset("core.bare")
		cfg.Set("core.logAllRefUpdates", "true")
		changed = true
	}
	if remoteName != "" && remoteURL != "" {
		cfg.Set("remote."+remoteName+".url", v.RemoteURL)
		cfg.Set("remote."+remoteName+".fetch", "+refs/heads/*:refs/remotes/"+remoteName+"/*")
		changed = true
	}
	if changed {
		err = cfg.Save(v.configFile())
	}
	return err
}

func (v *Repository) setAlternates(reference string) {
	var err error

	if reference != "" {
		// create file: objects/info/alternates
		altFile := filepath.Join(v.Path, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)
		var f *os.File
		f, err = os.OpenFile(altFile, os.O_CREATE|os.O_RDWR, 0644)
		defer f.Close()
		if err == nil {
			relPath := filepath.Join(reference, "objects")
			relPath, err = filepath.Rel(filepath.Join(v.Path, "objects"), relPath)
			if err == nil {
				_, err = f.WriteString(relPath + "\n")
			}
			if err != nil {
				log.Errorf("fail to set info/alternates on %s: %s", v.Path, err)
			}
		}
	}
}

// GitConfigRemoteURL returns remote url in git config
func (v *Repository) GitConfigRemoteURL(name string) string {
	return v.Config().Get("remote." + name + ".url")
}

// Init runs git-init on repository
func (v *Repository) Init(remoteName, remoteURL, referenceGitDir string) error {
	var err error

	if !v.Exists() {
		_, err = git.PlainInit(v.Path, true)
		if err != nil {
			return err
		}
		v.initMissing()
	}

	if remoteName != "" && remoteURL != "" {
		if !strings.HasSuffix(remoteURL, ".git") {
			remoteURL += ".git"
		}
		u := v.GitConfigRemoteURL(remoteName)
		if u != remoteURL {
			err = v.setRemote(remoteName, remoteURL)
			if err != nil {
				return err
			}
		}
	}

	if referenceGitDir != "" {
		v.setAlternates(referenceGitDir)
	}

	// TODO: Link hooks files in ../hooks/ dir to repository's hook dir.
	// TODO: Only copy 'commit-msg' hook, when: 1. gerrit mode, 2. defined v.Remote.Review

	return nil
}

func (v *Repository) initMissing() error {
	var err error

	if _, err = os.Stat(v.Path); err != nil {
		return err
	}

	dirs := []string{
		"hooks",
		"branches",
		"hooks",
		"info",
		"refs",
	}
	files := map[string]string{
		"description": fmt.Sprintf("Repository: %s, path: %s\n", v.Name, v.Path),
		"config":      "[core]\n\trepositoryformatversion = 0\n",
		"HEAD":        "ref: refs/heads/master\n",
	}

	for _, dir := range dirs {
		dir = filepath.Join(v.Path, dir)
		if _, err = os.Stat(dir); err == nil {
			continue
		}
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	for file, content := range files {
		file = filepath.Join(v.Path, file)
		if _, err = os.Stat(file); err == nil {
			continue
		}
		f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		f.WriteString(content)
		f.Close()
	}

	if !v.IsBare {
		cfg := v.Config()
		cfg.Unset("core.bare")
		cfg.Set("core.logAllRefUpdates", "true")
		cfg.Save(v.configFile())
	}

	return nil
}

// InitByLink will init repository by attaching other repository
func (v *Repository) InitByLink(remoteName, remoteURL string, repo *Repository) error {
	var err error

	if !repo.Exists() {
		return fmt.Errorf("attach a non-exist repo: %s", repo.Path)
	}
	repo.initMissing()

	err = os.MkdirAll(v.Path, 0755)
	if err != nil {
		return err
	}

	items := []string{
		"objects",
		"description",
		"info",
		"hooks",
		"svn",
		"rr-cache",
	}
	for _, item := range items {
		source := filepath.Join(repo.Path, item)
		target := filepath.Join(v.Path, item)
		if _, err = os.Stat(source); err != nil {
			continue
		}
		relpath, err := filepath.Rel(v.Path, source)
		if err != nil {
			relpath = source
		}
		err = os.Symlink(relpath, target)
		if err != nil {
			break
		}
	}
	v.initMissing()

	if remoteName != "" && remoteURL != "" {
		if !strings.HasSuffix(remoteURL, ".git") {
			remoteURL += ".git"
		}
		u := v.GitConfigRemoteURL(remoteName)
		if u != remoteURL {
			err = v.setRemote(remoteName, remoteURL)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Repository) isUnborn() bool {
	repo := v.Raw()
	if repo == nil {
		return false
	}
	_, err := repo.Head()
	return err != nil
}

func (v *Repository) fetchRefSpecs() []string {
	if len(v.RefSpecs) > 0 {
		return v.RefSpecs
	}

	return DefaultRefSpecs
}

// HasAlternates checks if repository has defined alternates
func (v *Repository) HasAlternates() bool {
	altFile := filepath.Join(v.Path, "objects", "info", "alternates")
	finfo, err := os.Stat(altFile)
	if err != nil {
		return false
	}
	if finfo.Size() == 0 {
		return false
	}
	return true
}

func (v *Repository) applyCloneBundle() {
	// TODO: download and clone from bundle file
}

// GetHead returns current branch name
func (v Repository) GetHead() string {
	f, err := os.Open(filepath.Join(v.Path, "HEAD"))
	if err != nil {
		return ""
	}
	defer f.Close()
	head, err := bufio.NewReader(f).ReadString('\n')
	if err != nil {
		return ""
	}
	if strings.HasPrefix(head, "ref:") {
		head = strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
		return head
	}
	return ""
}

// IsRebaseInProgress checks whether is in middle of a rebase.
func (v Repository) IsRebaseInProgress() bool {
	return path.Exists(filepath.Join(v.Path, "rebase-apply")) ||
		path.Exists(filepath.Join(v.Path, "rebase-merge")) ||
		path.Exists(filepath.Join(v.Path, ".dotest"))
}

// RevisionIsValid returns true if revision can be resolved
func (v Repository) RevisionIsValid(revision string) bool {
	raw := v.Raw()

	if raw == nil {
		return false
	}
	if _, err := raw.ResolveRevision(plumbing.Revision(revision)); err == nil {
		return true
	}
	return false
}

// Revlist works like rev-list
// TODO: Hack go-git plumbing/revlist package to replace git exec
func (v Repository) Revlist(args ...string) ([]string, error) {
	result := []string{}
	cmdArgs := []string{
		"git",
		"rev-list",
	}

	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = v.Path
	cmd.Stdin = nil
	cmd.Stderr = nil
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	r := bufio.NewReader(out)
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			result = append(result, line)
		}
		if err != nil {
			break
		}
	}

	if err = cmd.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// RemoteTrackBranch gets remote tracking branch
func (v Repository) RemoteTrackBranch(branch string) string {
	if branch == "" {
		branch = v.GetHead()
	}
	if branch == "" {
		return ""
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	cfg := v.Config()
	return cfg.Get("branch." + branch + ".merge")
}

// Fetch runs git-fetch on repository
func (v *Repository) Fetch(remote string, o *FetchOptions) error {
	var (
		err           error
		hasAlternates bool
	)

	if v.isUnborn() && v.Reference != "" && path.IsGitDir(v.Reference) {
		hasAlternates = true

		altFile := filepath.Join(v.Path, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)

		var f *os.File
		f, err = os.OpenFile(altFile, os.O_CREATE|os.O_RDWR, 0644)
		defer f.Close()
		if err == nil {
			target := filepath.Join(v.Reference, "objects")
			target, err = filepath.Rel(filepath.Join(v.Path, "objects"), target)
			if err != nil {
				target = filepath.Join(v.Reference, "objects")
			}
			_, err = f.WriteString(target + "\n")
		}
	} else if v.HasAlternates() {
		hasAlternates = true
	}

	if o.CloneBundle && !hasAlternates {
		v.applyCloneBundle()
	}

	if v.RemoteURL == "" {
		return fmt.Errorf("don't know where to fetch repo %s from remote %s", v.Name, remote)
	}

	if v.Revision == "" {
		v.Revision = v.RemoteTrackBranch("")
		if v.Revision == "" {
			log.Warnf("cannot get tracking branch for project '%s'", v.Name)
			v.Revision = "master"
		}
	}

	isSha := IsSha(v.Revision)
	isTag := IsTag(v.Revision)

	if o.OptimizedFetch && isSha && v.RevisionIsValid(v.Revision) {
		return nil
	}

	if o.Mirror && o.Depth > 0 {
		o.Depth = 0
	}
	if o.Depth > 0 {
		o.CurrentBranchOnly = true
	}
	if o.CurrentBranchOnly {
		if isSha || isTag {
			if v.RevisionIsValid(v.Revision) {
				return nil
			}
		}
	}

	cmdArgs := []string{
		GIT,
		"fetch",
	}

	if o.Depth > 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--depth=%d", o.Depth))
	} else if path.Exist(filepath.Join(v.Path, "shallow")) {
		cmdArgs = append(cmdArgs, "--unshallow")
	}

	if o.Quiet {
		cmdArgs = append(cmdArgs, "--quiet")

	}

	if o.NoTags || o.Depth > 0 {
		cmdArgs = append(cmdArgs, "--no-tags")
	} else {
		cmdArgs = append(cmdArgs, "--tags")
	}

	if o.Prune {
		cmdArgs = append(cmdArgs, "--prune")

	}

	if o.Submodules {
		cmdArgs = append(cmdArgs, "--recurse-submodules=on-demand")
	}

	cmdArgs = append(cmdArgs, v.RemoteURL)
	if o.CurrentBranchOnly {
		if isSha {
			cmdArgs = append(cmdArgs, v.Revision)
		} else if isTag {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+%s:%s", v.Revision, v.Revision))
		} else if strings.HasPrefix(v.Revision, "refs/heads/") || !strings.HasPrefix(v.Revision, "refs/") {
			branch := strings.TrimPrefix(v.Revision, "refs/heads/")
			if v.IsBare {
				cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/%s:refs/heads/%s", branch, branch))
			} else {
				cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, v.Remote, branch))
			}
		} else {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+%s:%s", v.Revision, v.Revision))
		}
	} else {
		if v.IsBare {
			cmdArgs = append(cmdArgs, "+refs/heads/*:refs/heads/*")
		} else {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", v.Remote))
		}
	}

	err = executeCommandIn(v.Path, cmdArgs)
	if err != nil {
		return err
	}

	if hasAlternates && v.Settings.Dissociate {
		cmdArgs = []string{
			GIT,
			"repack",
			"-a",
			"-d",
		}
		err = executeCommandIn(v.Path, cmdArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Raw returns go-git repository object
func (v *Repository) Raw() *git.Repository {
	var err error

	if v.raw != nil {
		return v.raw
	}

	v.raw, err = git.PlainOpen(v.Path)
	if err != nil {
		log.Errorf("cannot open git repo '%s': %s", v.Path, err)
		return nil
	}
	return v.raw
}

func (v *Repository) configFile() string {
	return filepath.Join(v.Path, "config")
}

// Config returns git config file parser
func (v *Repository) Config() goconfig.GitConfig {
	cfg, err := goconfig.Load(v.configFile())
	if err != nil && err != goconfig.ErrNotExist {
		log.Fatalf("fail to load config: %s: %s", v.configFile(), err)
	}
	if cfg == nil {
		cfg = goconfig.NewGitConfig()
	}
	return cfg
}

// SaveConfig will save config to git config file
func (v *Repository) SaveConfig(cfg goconfig.GitConfig) error {
	if cfg == nil {
		cfg = goconfig.NewGitConfig()
	}
	return cfg.Save(v.configFile())
}

// DeleteBranch deletes a branch
func (v Repository) DeleteBranch(branch string) error {
	// TODO: go-git fail to delete a branch
	// TODO: return v.Raw().DeleteBranch(branch)
	if IsHead(branch) {
		branch = strings.TrimPrefix(branch, config.RefsHeads)
	}
	cmdArgs := []string{
		GIT,
		"branch",
		"-D",
		branch,
		"--",
	}
	return executeCommandIn(v.Path, cmdArgs)
}
