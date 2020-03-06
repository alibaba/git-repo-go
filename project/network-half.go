package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alibaba/git-repo-go/file"
	"github.com/alibaba/git-repo-go/path"
	log "github.com/jiangxin/multi-log"
)

// FetchOptions is options for git fetch.
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

// Fetch runs git-fetch on repository.
func (v *Repository) Fetch(remote string, o *FetchOptions) error {
	var (
		err           error
		hasAlternates bool
		revision      = v.Revision
	)

	if v.isUnborn() && v.Reference != "" && path.IsGitDir(v.Reference) {
		hasAlternates = true

		altFile := filepath.Join(v.GitDir, "objects", "info", "alternates")
		os.MkdirAll(filepath.Dir(altFile), 0755)

		var f *os.File
		f, err = file.New(altFile).OpenCreateRewrite()
		defer f.Close()
		if err == nil {
			target := filepath.Join(v.Reference, "objects")
			target, err = filepath.Rel(filepath.Join(v.GitDir, "objects"), target)
			if err != nil {
				target = filepath.Join(v.Reference, "objects")
			}
			_, err = f.WriteString(target + "\n")
			log.Debugf("%slinked with reference repo: %s", v.Prompt(), target)
		}
		if err != nil {
			log.Warnf("%sfail to link with reference repo: %s", v.Prompt(), err)
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

	if revision == "" {
		revision = v.TrackBranch("")
		if revision == "" {
			log.Warnf("cannot get tracking branch for project '%s'", v.Name)
			revision = "master"
		}
	}

	isSha := IsSha(revision)
	isTag := IsTag(revision)

	if o.OptimizedFetch && isSha && v.RevisionIsValid(revision) {
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
			if v.RevisionIsValid(revision) {
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
	} else if path.Exist(filepath.Join(v.RepoDir(), "shallow")) {
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
			cmdArgs = append(cmdArgs, revision)
		} else if isTag {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+%s:%s", revision, revision))
		} else if strings.HasPrefix(revision, "refs/heads/") || !strings.HasPrefix(revision, "refs/") {
			branch := strings.TrimPrefix(revision, "refs/heads/")
			if v.IsBare {
				cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/%s:refs/heads/%s", branch, branch))
			} else {
				cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, v.RemoteName, branch))
			}
		} else {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+%s:%s", revision, revision))
		}
	} else {
		if v.IsBare {
			cmdArgs = append(cmdArgs, "+refs/heads/*:refs/heads/*")
		} else {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", v.RemoteName))
		}
	}
	log.Debugf("%sfetching using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))

	err = executeCommandIn(v.RepoDir(), cmdArgs)
	if err != nil {
		return fmt.Errorf("fail to fetch project '%s': %s", v.Name, err)
	}

	if hasAlternates && v.Settings.Dissociate {
		cmdArgs = []string{
			GIT,
			"repack",
			"-a",
			"-d",
		}
		log.Debugf("%srepacking using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))

		err = executeCommandIn(v.RepoDir(), cmdArgs)
		if err != nil {
			return fmt.Errorf("fail to repack '%s': %s", v.Name, err)
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
	log.Debugf("%sfetching archive in %s: %s", v.Prompt(), v.TopDir(), strings.Join(cmdArgs, " "))

	return executeCommandIn(v.TopDir(), cmdArgs)
}

func (v *Project) extractArchive(tarpath string) error {
	cmdArgs := []string{
		"tar",
		"-x",
		"-f",
		tarpath,
	}
	log.Debugf("%sextracting archive in %s: %s", v.Prompt(), v.TopDir(), strings.Join(cmdArgs, " "))

	return executeCommandIn(v.TopDir(), cmdArgs)
}

// SyncNetworkHalf starts to fetch from remote repository.
func (v *Project) SyncNetworkHalf(o *FetchOptions) error {
	var err error

	if o == nil {
		o = &FetchOptions{}
	}

	remoteURL, err := v.GetRemoteURL()
	if err != nil {
		return err
	}
	if o.Archive && !v.IsMetaProject() {
		if strings.HasPrefix(remoteURL, "http://") ||
			strings.HasPrefix(remoteURL, "https://") {
			return fmt.Errorf("%s: Cannot fetch archives from http/https remotes", v.Name)
		}

		tarpath := strings.Replace(v.Name, "/", "_", -1)
		tarpath += ".tar"
		err = v.fetchArchive(tarpath)
		if err != nil {
			return fmt.Errorf("fail to fetch tarball %s: %s", tarpath, err)
		}
		err = v.extractArchive(tarpath)
		if err != nil {
			return fmt.Errorf("fail to extract tarball %s: %s", tarpath, err)
		}
		err = os.Remove(filepath.Join(v.TopDir(), tarpath))
		if err != nil {
			return fmt.Errorf("cannot remove tarball %s: %s", tarpath, err)
		}
		return v.CopyAndLinkFiles()
	}

	if !v.Repository.Exists() {
		// Initial repository
		v.GitInit()
	}
	return v.Repository.Fetch(v.RemoteName, o)
}
