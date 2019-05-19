package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/multi-log"
)

// FetchOptions is options for git fetch
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
		v.Revision = v.TrackBranch("")
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
				cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", branch, v.RemoteName, branch))
			}
		} else {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+%s:%s", v.Revision, v.Revision))
		}
	} else {
		if v.IsBare {
			cmdArgs = append(cmdArgs, "+refs/heads/*:refs/heads/*")
		} else {
			cmdArgs = append(cmdArgs, fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", v.RemoteName))
		}
	}

	err = executeCommandIn(v.Path, cmdArgs)
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
		err = executeCommandIn(v.Path, cmdArgs)
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

	return executeCommandIn(v.RepoRoot(), cmdArgs)
}

func (v *Project) extractArchive(tarpath string) error {
	cmdArgs := []string{
		"tar",
		"-x",
		"-f",
		tarpath,
	}

	return executeCommandIn(v.RepoRoot(), cmdArgs)
}

// SyncNetworkHalf will fetch from remote repository
func (v *Project) SyncNetworkHalf(o *FetchOptions) error {
	var err error

	if o == nil {
		o = &FetchOptions{}
	}

	// Initial repository
	v.GitInit()

	remoteURL, err := v.GetRemoteURL()
	if err != nil {
		return err
	}
	if o.Archive && !v.IsMetaProject() {
		if strings.HasPrefix(remoteURL, "http://") ||
			strings.HasPrefix(remoteURL, "https://") {
			return fmt.Errorf("%s: Cannot fetch archives from http/https remotes", v.Name)
		}

		tarpath := strings.ReplaceAll(v.Name, "/", "_")
		tarpath += ".tar"
		err = v.fetchArchive(tarpath)
		if err != nil {
			return fmt.Errorf("fail to fetch tarball %s: %s", tarpath, err)
		}
		err = v.extractArchive(tarpath)
		if err != nil {
			return fmt.Errorf("fail to extract tarball %s: %s", tarpath, err)
		}
		err = os.Remove(filepath.Join(v.RepoRoot(), tarpath))
		if err != nil {
			return fmt.Errorf("cannot remove tarball %s: %s", tarpath, err)
		}
		return v.CopyAndLinkFiles()
	}

	if v.WorkRepository != nil {
		err = v.WorkRepository.Fetch(v.RemoteName, o)
		if err != nil {
			return err
		}
	}
	log.Debugf("WorkRepository of project '%s' is nil, sync-network-half failed", v.Name)
	return nil
}
