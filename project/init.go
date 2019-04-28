package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

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
			v.WorkRepository.Init(v.RemoteName, remoteURL, referenceGitDir)
		} else {
			v.WorkRepository.InitByLink(v.RemoteName, remoteURL, v.ObjectRepository)
		}
	}

	// TODO: install hooks
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
