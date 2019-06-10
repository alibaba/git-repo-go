package project

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Macros for project package
const (
	GIT = "git"
)

// Repository has repository related operations
type Repository struct {
	Name       string // Project name
	Path       string // Repository real path
	IsBare     bool
	RemoteURL  string
	Reference  string // Alternate repository
	RemoteName string // Project RemoteName field from manifest xml
	Revision   string // Projeect Revision from manifest xml
	Settings   *RepoSettings
	raw        *git.Repository
}

// Exists checks repository layout
func (v Repository) Exists() bool {
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

func (v Repository) setAlternates(reference string) {
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
func (v Repository) GitConfigRemoteURL(name string) string {
	return v.Config().Get("remote." + name + ".url")
}

func (v Repository) isUnborn() bool {
	repo := v.Raw()
	if repo == nil {
		return false
	}
	_, err := repo.Head()
	return err != nil
}

// HasAlternates checks if repository has defined alternates
func (v Repository) HasAlternates() bool {
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

func (v Repository) applyCloneBundle() {
	// TODO: download and clone from bundle file
}

// GetHead returns current branch name
func (v Repository) GetHead() string {
	r := v.Raw()
	if r == nil {
		return ""
	}

	// Not checkout yet
	head, err := r.Head()
	if head == nil || err != nil {
		return ""
	}

	headName := head.Name().String()
	if headName == "HEAD" {
		return ""
	}
	return headName
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

// Raw returns go-git repository object
func (v Repository) Raw() *git.Repository {
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

func (v Repository) configFile() string {
	return filepath.Join(v.Path, "config")
}

// Config returns git config file parser
func (v Repository) Config() goconfig.GitConfig {
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
