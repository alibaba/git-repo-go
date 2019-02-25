package project

import (
	"fmt"
	"os"
	"path/filepath"

	"code.alibaba-inc.com/force/git-repo/cap"
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
	ProjectName   string // Project name
	Path          string
	RefSpecs      []string
	OnceReference string // First clone/fetch, can use this as a reference
	IsBare        bool
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

// Init runs git-init on repository
func (v *Repository) Init(remoteName, remoteURL, referenceGitDir string) error {
	var err error

	if !v.Exists() {
		_, err = git.PlainInit(v.Path, true)
		if err != nil {
			return err
		}
	}

	cfg := v.Config()
	changed := false
	if !v.IsBare {
		cfg.Unset("core.bare")
		cfg.Set("core.logAllRefUpdates", "true")
		changed = true
	}
	if remoteName != "" && remoteURL != "" {
		cfg.Set("remote."+remoteName+".url", remoteURL)
		changed = true
	}
	if changed {
		cfg.Save(v.configFile())
	}

	if referenceGitDir != "" {
		// create file: objects/info/alternates
		altFile := filepath.Join(v.Path, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)
		var f *os.File
		f, err = os.OpenFile(altFile, os.O_CREATE|os.O_RDWR, 0644)
		defer f.Close()
		if err == nil {
			relPath := filepath.Join(referenceGitDir, "objects")
			relPath, err = filepath.Rel(filepath.Join(v.Path, "objects"), relPath)
			if err == nil {
				_, err = f.WriteString(relPath + "\n")
			}
			if err != nil {
				log.Errorf("fail to set info/alternates on %s: %s", v.Path, err)
			}
		}
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
	}
	files := map[string]string{
		"description": fmt.Sprintf("Repository: %s, path: %s\n", v.Name(), v.Path),
		"config":      "[core]\n\trepositoryformatversion = 0\n",
		"HEAD":        "ref: refs/heads/master\n",
	}

	for _, dir := range dirs {
		if err = os.MkdirAll(filepath.Join(v.Path, dir), 0755); err != nil {
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

	return nil
}

// InitByAttach will init repository by attaching other repository
func (v *Repository) InitByAttach(repo *Repository) error {
	var err error

	if !repo.Exists() {
		return fmt.Errorf("attach a non-exist repo: %s", repo.Path)
	}
	repo.initMissing()

	err = os.MkdirAll(v.Path, 0755)
	if err != nil {
		return err
	}

	if cap.Symlink() {
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
			err = os.Symlink(source, target)
			if err != nil {
				break
			}
		}
		v.initMissing()
	} else {
		err = v.Init("", "", "")
		if err != nil {
			return err
		}
		// create file: objects/info/alternates
		altFile := filepath.Join(v.Path, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)
		var f *os.File
		f, err = os.OpenFile(altFile, os.O_CREATE|os.O_RDWR, 0644)
		defer f.Close()
		if err == nil {
			relPath := filepath.Join(repo.Path, "objects")
			relPath, err = filepath.Rel(filepath.Join(v.Path, "objects"), relPath)
			if err != nil {
				relPath = filepath.Join(repo.Path, "objects")
			}
			_, err = f.WriteString(relPath + "\n")
		}
	}
	if err != nil {
		return fmt.Errorf("fail to create alternates file: %s", err)
	}
	return nil
}

func (v *Repository) isUnborn() bool {
	repo := v.Raw()
	_, err := repo.Head()
	return err != nil
}

func (v *Repository) fetchRefSpecs() []string {
	if len(v.RefSpecs) > 0 {
		return v.RefSpecs
	}

	return DefaultRefSpecs
}

// Fetch runs git-fetch on repository
func (v *Repository) Fetch(remote string) error {
	var err error

	if v.isUnborn() && v.OnceReference != "" {
		// fetch from reference repo first
		if path.IsGitDir(v.OnceReference) {
			// fetch first
			cmdArgs := []string{
				GIT,
				"fetch",
				v.OnceReference,
			}
			cmdArgs = append(cmdArgs, v.fetchRefSpecs()...)
			err = executeCommandIn(v.Path, cmdArgs)
			if err != nil {
				log.Errorf("fail to fetch %s from reference: %s", v.Name(), err)
			}
		}
	}

	remoteURL := v.Config().Get("remote." + remote + ".url")
	if remoteURL == "" {
		return fmt.Errorf("don't know where to fetch repo %s from remote %s", v.Name(), remote)
	}

	cmdArgs := []string{
		GIT,
		"fetch",
		"--prune",
		remoteURL,
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
		log.Fatalf("cannot open repo: %s", v.Path)
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
