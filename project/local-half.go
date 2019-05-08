package project

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/multi-log"
)

// CheckoutOptions is options for git fetch
type CheckoutOptions struct {
	RepoSettings

	Quiet      bool
	DetachHead bool
}

// IsClean checks whether repository dir is clean
func IsClean(dir string) (bool, error) {
	if !path.Exists(dir) {
		return false, fmt.Errorf("dir %s does not exist", dir)
	}

	status := []string{}
	cmdArgs := []string{
		GIT,
		"status",
		"-uno",
		"--porcelain",
		"--",
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.Stderr = nil
	out, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	if err = cmd.Start(); err != nil {
		return false, err
	}

	r := bufio.NewReader(out)
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			status = append(status, line)
		}
		if err != nil {
			break
		}
	}

	if err = cmd.Wait(); err != nil {
		return false, err
	}

	if len(status) == 0 {
		return true, nil
	}

	return false, nil
}

// IsClean indicates worktree is clean or dirty.
// TODO: go-git failed with invalid checksum
func (v Project) IsClean() (bool, error) {
	return IsClean(v.WorkDir)
}

// CheckoutRevision runs git checkout
func (v Project) CheckoutRevision(args ...string) error {
	cmdArgs := []string{
		GIT,
		"checkout",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// HardReset runs git reset --hard
func (v Project) HardReset(args ...string) error {
	cmdArgs := []string{
		GIT,
		"reset",
		"--hard",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// Rebase runs git rebase
func (v Project) Rebase(args ...string) error {
	cmdArgs := []string{
		GIT,
		"rebase",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// FastForward runs git merge
func (v Project) FastForward(args ...string) error {
	cmdArgs := []string{
		GIT,
		"merge",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// SubmoduleUpdate runs git submodule update
func (v Project) SubmoduleUpdate(args ...string) error {
	cmdArgs := []string{
		GIT,
		"submodule",
		"update",
		"--init",
		"--recursive",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// SyncLocalHalf will checkout/rebase branch.
func (v Project) SyncLocalHalf(o *CheckoutOptions) error {
	var (
		err error
	)

	err = v.PrepareWorkdir()
	if err != nil {
		return err
	}

	if v.Revision == "" {
		return nil
	}

	// Remove obsolete refs/published/ references
	v.CleanPublishedCache()

	// Get revision id of already fetch v.Revision, will checkout to revid later.
	revid, err := v.ResolveRemoteTracking(v.Revision)
	if err != nil || revid == "" {
		return fmt.Errorf("cannot checkout, invalid remote tracking branch '%s': %s",
			v.Revision,
			err)
	}

	// Read current branch to 'branch' and parsed revision to 'headid'
	// If repository is in detached head mode, or has invalid HEAD, branch is empty.
	branch := ""
	track := ""
	head := v.GetHead()
	headid, err := v.ResolveRevision(head)
	if err == nil && headid != "" {
		if IsHead(head) {
			branch = strings.TrimPrefix(head, config.RefsHeads)
		}
	}

	// We have a branch, check whether tracking branch is set properly.
	if branch != "" {
		track = v.RemoteTrackBranch(branch)
	}

	log.Debugf("Fetch (project, %s, head: %s, branch: %s, track: %s, headid: %s, revid: %s, revision: %s)",
		v.Name, head, branch, track, headid, revid, v.Revision)

	PostUpdate := func(update bool) error {
		var err error

		if branch != "" && track != v.Revision {
			if v.Revision != "" {
				if track == "" {
					log.Notef("manifest switched to %s", v.Revision)
				} else {
					log.Notef("manifest switched %s...%s", track, v.Revision)
				}
			} else {
				log.Notef("manifest no longer tracks %s", track)
			}

			// Update remote tracking, or delete tracking if v.Revision is empty
			v.UpdateBranchTracking(branch, v.RemoteName, v.Revision)
		}

		if update && o.Submodules {
			err = v.SubmoduleUpdate()
			if err != nil {
				return err
			}
		}
		return v.CopyAndLinkFiles()
	}

	// Currently on a detached HEAD.  The user is assumed to
	// not have any local modifications worth worrying about.
	if branch == "" || o.DetachHead {
		if v.IsRebaseInProgress() {
			return fmt.Errorf("prior sync failed; rebase still in progress")
		}

		if headid == revid {
			return PostUpdate(false)
		}

		if headid != "" {
			localChanges, err := v.Revlist(headid, "--not", revid)
			if err != nil {
				log.Warnf("rev-list failed: %s", err)
			}
			if len(localChanges) > 0 {
				log.Notef("discarding %d commits", len(localChanges))
			}
		}

		log.Debugf("detached head, force checkout: %s", revid)
		err = v.CheckoutRevision("--force", revid)
		if err != nil {
			return err
		}

		return PostUpdate(true)
	}

	// No need to checkout
	if headid == revid {
		return PostUpdate(false)
	}

	// No track, no loose.
	if track == "" {
		log.Notef("leaving %s; does not track upstream", branch)
		if branch != "" {
			err = v.HardReset(revid)
		} else {
			err = v.CheckoutRevision("--force", revid)
		}
		if err != nil {
			return err
		}
		return PostUpdate(true)
	}

	remoteChanges, err := v.Revlist(revid, "--not", headid)
	if err != nil {
		log.Errorf("rev-list failed: %s", err)
	}

	// No remote changes, no update.
	if len(remoteChanges) == 0 {
		log.Debugf("no remote changes found for project %s", v.Name)
		return PostUpdate(false)
	}

	pubid := v.PublishedRevision(branch)
	// Local branched is published.
	if pubid != "" {
		notMerged, err := v.Revlist(pubid, "--not", revid)
		if err != nil {
			return fmt.Errorf("fail to check publish status for branch '%s': %s",
				branch,
				err)
		}
		// Has unpublished changes, fail to update.
		if len(notMerged) > 0 {
			log.Debugf("has %d unpublished commit(s) for project %s", len(notMerged), v.Name)
			if len(remoteChanges) > 0 {
				log.Errorf("branch %s is published (but not merged) and is now "+
					"%d commits behind", branch, len(remoteChanges))
			}
			return fmt.Errorf("branch %s is published (but not merged)", branch)
		}
		// Since last published, no other local changes.
		if pubid == headid {
			log.Debugf("all local commits are published for project %s", v.Name)
			err = v.FastForward(revid)
			if err != nil {
				return err
			}
			return PostUpdate(true)
		}
	}

	// Failed if worktree is dirty.
	if ok, _ := v.IsClean(); !ok {
		return fmt.Errorf("worktree of %s is dirty, checkout failed", v.Name)
	}

	localChanges, err := v.Revlist(headid, "--not", revid)
	if err != nil {
		log.Warnf("rev-list for local changes failed: %s", err)
	}

	if v.IsRebase() {
		err = v.Rebase(revid)
		if err != nil {
			return err
		}
	} else if len(localChanges) > 0 {
		err = v.HardReset(revid)
		if err != nil {
			return err
		}
	} else {
		err = v.FastForward(revid)
		if err != nil {
			return err
		}
	}

	return PostUpdate(true)
}

// CopyFile copy files from src to dest
func (v Project) CopyFile(src, dest string) error {
	srcAbs := filepath.Clean(filepath.Join(v.WorkDir, src))
	destAbs := filepath.Clean(filepath.Join(v.RepoRoot(), dest))

	if !strings.HasPrefix(srcAbs, v.RepoRoot()) {
		return fmt.Errorf("fail to copy file, src file '%s' beyond repo root '%s'", src, v.RepoRoot())
	}

	if !strings.HasPrefix(destAbs, v.RepoRoot()) {
		return fmt.Errorf("fail to copy file, dest file '%s' beyond repo root '%s'", dest, v.RepoRoot())
	}

	finfo, err := os.Stat(srcAbs)
	if err != nil {
		return nil
	}

	if !path.Exists(filepath.Dir(destAbs)) {
		os.MkdirAll(filepath.Dir(destAbs), 0755)
	}

	srcFile, err := os.Open(srcAbs)
	if err != nil {
		return fmt.Errorf("fail to open '%s': %s", srcAbs, err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destAbs, os.O_RDWR|os.O_CREATE, finfo.Mode())
	if err != nil {
		return fmt.Errorf("fail to open '%s': %s", destAbs, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("fail to copy file: %s", err)
	}
	return nil
}

// LinkFile copy files from src to dest
func (v Project) LinkFile(src, dest string) error {
	srcAbs := filepath.Clean(filepath.Join(v.WorkDir, src))
	destAbs := filepath.Clean(filepath.Join(v.RepoRoot(), dest))

	if !strings.HasPrefix(srcAbs, v.RepoRoot()) {
		return fmt.Errorf("fail to copy file, src file '%s' beyond repo root '%s'", src, v.RepoRoot())
	}

	if !strings.HasPrefix(destAbs, v.RepoRoot()) {
		return fmt.Errorf("fail to copy file, dest file '%s' beyond repo root '%s'", dest, v.RepoRoot())
	}

	_, err := os.Stat(srcAbs)
	if err != nil {
		return nil
	}

	destDir := filepath.Dir(destAbs)
	if !path.Exists(destDir) {
		os.MkdirAll(destDir, 0755)
	}
	srcRel, err := filepath.Rel(destDir, srcAbs)
	if err != nil {
		srcRel = srcAbs
	}
	if path.Exists(destAbs) {
		os.Remove(destAbs)
	}
	if cap.CanSymlink() {
		return os.Symlink(srcRel, destAbs)
	}
	return os.Link(srcRel, destAbs)
}

// CopyAndLinkFiles copies and links files
func (v Project) CopyAndLinkFiles() error {
	var (
		err  error
		errs = []string{}
	)

	for _, f := range v.CopyFiles {
		err = v.CopyFile(f.Src, f.Dest)
		if err != nil {
			errs = append(errs,
				fmt.Sprintf("fail to copy file from %s to %s: %s", f.Src, f.Dest, err))
		}
	}

	for _, f := range v.LinkFiles {
		err = v.LinkFile(f.Src, f.Dest)
		if err != nil {
			errs = append(errs,
				fmt.Sprintf("fail to link file from %s to %s: %s", f.Src, f.Dest, err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
