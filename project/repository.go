package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
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
	ProjectName    string // Project name
	Path           string
	RefSpecs       []string
	LocalReference string // First clone/fetch, can use this as a reference
	IsBare         bool
	RemoteURL      string
}

// Name returns repository name
func (v *Repository) Name() string {
	if v.ProjectName != "" {
		return v.ProjectName
	}
	return filepath.Dir(v.Path)
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
		"description": fmt.Sprintf("Repository: %s, path: %s\n", v.Name(), v.Path),
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

// Fetch runs git-fetch on repository
func (v *Repository) Fetch(remote string, o *config.FetchOptions) error {
	var (
		err           error
		hasAlternates bool
	)

	if v.isUnborn() && v.LocalReference != "" && path.IsGitDir(v.LocalReference) {
		hasAlternates = true

		altFile := filepath.Join(v.Path, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)

		var f *os.File
		f, err = os.OpenFile(altFile, os.O_CREATE|os.O_RDWR, 0644)
		defer f.Close()
		if err == nil {
			target := filepath.Join(v.LocalReference, "objects")
			target, err = filepath.Rel(filepath.Join(v.Path, "objects"), target)
			if err != nil {
				target = filepath.Join(v.LocalReference, "objects")
			}
			_, err = f.WriteString(target + "\n")
		}
	} else if v.HasAlternates() {
		hasAlternates = true
	}

	if o.CloneBundle && !hasAlternates {
		v.applyCloneBundle()
	}

	// if o.GetCloneDepth()
	/*
		need_to_fetch = !(o.GetOptimizedFetch() &&
			(config.CommitIDPattern.MatchString(self.revisionExpr) &&
				self._CheckForImmutableRevision()))
	*/

	if v.RemoteURL == "" {
		return fmt.Errorf("don't know where to fetch repo %s from remote %s", v.Name(), remote)
	}

	cmdArgs := []string{
		GIT,
		"fetch",
		"--prune",
		v.RemoteURL,
	}
	cmdArgs = append(cmdArgs, v.fetchRefSpecs()...)
	err = executeCommandIn(v.Path, cmdArgs)
	if err != nil {
		return err
	}
	return nil
}

// Raw returns go-git repository object
func (v *Repository) Raw() *git.Repository {
	repo, err := git.PlainOpen(v.Path)
	if err != nil {
		return nil
	}
	return repo
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
