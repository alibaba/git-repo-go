package project

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const (
	// GIT is git command name
	GIT = config.GIT
)

// Repository has repository related operations.
type Repository struct {
	manifest.Project

	DotGit        string // Path to worktree/.git
	GitDir        string // Project's bare repository inside .repo
	ObjectsGitDir string // Several projects may share the same repository

	IsBare    bool
	RemoteURL string
	Remotes   *RemoteMap
	Reference string // Alternate repository
	Settings  *RepoSettings
	raw       *git.Repository
}

// RepoDir returns git dir of the repository
func (v Repository) RepoDir() string {
	if path.IsDir(v.DotGit) {
		return v.DotGit
	}
	return v.GitDir
}

// ObjectsRepository returns repository which ObjectsGitDir points to
func (v Repository) ObjectsRepository() *Repository {
	if v.ObjectsGitDir == "" {
		return nil
	}

	return &Repository{
		Project: v.Project,

		DotGit:        "",
		GitDir:        v.ObjectsGitDir,
		ObjectsGitDir: "",

		IsBare:    true,
		RemoteURL: v.RemoteURL,
		Settings:  v.Settings,
		Remotes:   nil,
	}
}

// Exists checks repository layout.
func (v Repository) Exists() bool {
	return path.IsGitDir(v.GitDir)
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
		altFile := filepath.Join(v.GitDir, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)
		var f *os.File
		f, err = os.OpenFile(altFile, os.O_CREATE|os.O_RDWR, 0644)
		defer f.Close()
		if err == nil {
			relPath := filepath.Join(reference, "objects")
			relPath, err = filepath.Rel(filepath.Join(v.GitDir, "objects"), relPath)
			if err == nil {
				_, err = f.WriteString(relPath + "\n")
			}
			if err != nil {
				log.Errorf("fail to set info/alternates on %s: %s", v.GitDir, err)
			}
		}
	}
}

// GitConfigRemoteURL returns remote url in git config.
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

// HasAlternates checks if repository has defined alternates.
func (v Repository) HasAlternates() bool {
	altFile := filepath.Join(v.GitDir, "objects", "info", "alternates")
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
	gitDir := v.RepoDir()
	return path.Exist(filepath.Join(gitDir, "rebase-apply")) ||
		path.Exist(filepath.Join(gitDir, "rebase-merge")) ||
		path.Exist(filepath.Join(gitDir, ".dotest"))
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

// LastModified gets last modified time of a revision
func (v Repository) LastModified(revision string) string {
	raw := v.Raw()

	if raw == nil {
		return ""
	}
	obj, err := raw.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return ""
	}
	commit, err := raw.CommitObject(*obj)
	if err != nil {
		return ""
	}

	return commit.Committer.When.Format("Mon Jan 2 15:04:05 -0700 2006")
}

// Revlist works like rev-list.
// TODO: Hack go-git plumbing/revlist package to replace git exec
func (v Repository) Revlist(args ...string) ([]string, error) {
	result := []string{}
	cmdArgs := []string{
		"git",
		"rev-list",
	}

	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = v.RepoDir()
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

// Raw returns go-git repository object.
func (v Repository) Raw() *git.Repository {
	var (
		err error
	)

	if v.raw != nil {
		return v.raw
	}

	v.raw, err = git.PlainOpen(v.RepoDir())
	if err != nil {
		log.Errorf("cannot open git repo '%s': %s", v.RepoDir(), err)
		return nil
	}
	return v.raw
}

func (v Repository) configFile() string {
	return filepath.Join(v.RepoDir(), "config")
}

// SSHInfoCacheFile is filename used to cache proto settings.
func (v Repository) SSHInfoCacheFile() string {
	return filepath.Join(v.RepoDir(), "info", "sshinfo.cache")
}

// Config returns git config file parser.
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

// SaveConfig will save config to git config file.
func (v *Repository) SaveConfig(cfg goconfig.GitConfig) error {
	if cfg == nil {
		cfg = goconfig.NewGitConfig()
	}
	return cfg.Save(v.configFile())
}

// Prompt will show project path as prompt.
func (v Repository) Prompt() string {
	if v.Path == "." {
		return ""
	}
	return v.Path + "> "
}
