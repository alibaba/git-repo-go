package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/path"
	"gopkg.in/src-d/go-git.v4"
)

// IsRepoInitialized indicates repository is initialized or not.
func (v Project) IsRepoInitialized() bool {
	if v.ObjectsGitDir != "" {
		if !path.IsGitDir(v.ObjectsGitDir) {
			return false
		}
	}
	if !path.IsGitDir(v.GitDir) {
		return false
	}
	return true
}

// GitInit starts to init repositories.
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

	objectsRepo := v.ObjectsRepository()
	if objectsRepo != nil {
		objectsRepo.Init("", "", "")
		v.Repository.InitByLink(v.RemoteName, remoteURL, objectsRepo)
	} else {
		v.Repository.Init(v.RemoteName, remoteURL, referenceGitDir)
	}

	// TODO: install hooks
	return nil
}

func (v *Repository) initMissing() error {
	var err error

	if _, err = os.Stat(v.GitDir); err != nil {
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
		"description": fmt.Sprintf("Repository: %s, path: %s\n", v.Name, v.GitDir),
		"config":      "[core]\n\trepositoryformatversion = 0\n",
		"HEAD":        "ref: refs/heads/master\n",
	}

	for _, dir := range dirs {
		dir = filepath.Join(v.GitDir, dir)
		if _, err = os.Stat(dir); err == nil {
			continue
		}
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	for file, content := range files {
		file = filepath.Join(v.GitDir, file)
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

// Init runs git-init on repository.
func (v *Repository) Init(remoteName, remoteURL, referenceGitDir string) error {
	var err error

	if !v.Exists() {
		_, err = git.PlainInit(v.GitDir, true)
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

// InitByLink starts to init repository by attaching other repository.
func (v *Repository) InitByLink(remoteName, remoteURL string, repo *Repository) error {
	var err error

	if !repo.Exists() {
		return fmt.Errorf("attach a non-exist repo: %s", repo.GitDir)
	}
	repo.initMissing()

	err = os.MkdirAll(v.GitDir, 0755)
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
		source := filepath.Join(repo.GitDir, item)
		target := filepath.Join(v.GitDir, item)
		if _, err = os.Stat(source); err != nil {
			continue
		}
		relpath, err := filepath.Rel(v.GitDir, source)
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
