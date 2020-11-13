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

	"github.com/alibaba/git-repo-go/cap"
	"github.com/alibaba/git-repo-go/common"
	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/file"
	"github.com/alibaba/git-repo-go/helper"
	"github.com/alibaba/git-repo-go/path"
	log "github.com/jiangxin/multi-log"
)

// CheckoutOptions is options for git fetch.
type CheckoutOptions struct {
	RepoSettings

	Quiet      bool
	DetachHead bool
	IsManifest bool
}

// IsClean indicates git worktree is clean.
func IsClean(dir string) (bool, error) {
	if !path.Exist(dir) {
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

// IsClean indicates git worktree is clean.
// TODO: cannot use go-git, because it is incompatible with git new index format.
func (v Project) IsClean() bool {
	ok, err := IsClean(v.WorkDir)
	if err != nil {
		log.Warnf("%sfail to run IsClean: %s", v.Prompt(), err)
	}
	return ok
}

// CheckoutRevision runs git checkout.
func (v Project) CheckoutRevision(args ...string) error {
	cmdArgs := []string{
		GIT,
		"checkout",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	log.Debugf("%schecking out using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// HardReset runs git reset --hard.
func (v Project) HardReset(args ...string) error {
	cmdArgs := []string{
		GIT,
		"reset",
		"--hard",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	log.Debugf("%shard reset using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// Rebase runs git rebase.
func (v Project) Rebase(args ...string) error {
	cmdArgs := []string{
		GIT,
		"rebase",
	}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--")
	log.Debugf("%srebasing using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))
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
	log.Debugf("%sfastforward using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// SubmoduleUpdate runs git submodule update.
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
	log.Debugf("%ssubmodule update using command: %s", v.Prompt(), strings.Join(cmdArgs, " "))
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
		log.Debugf("%sRevision is empty, do nothing", v.Prompt())
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
	log.Debugf("%sremote tracking ref for %s: %s", v.Prompt(), v.Revision, revid)

	// Read current branch to 'branch' and parsed revision to 'headid'
	// If repository is in detached head mode, or has invalid HEAD, branch is empty.
	branch := ""
	track := ""
	head := v.GetHead()
	headid, err := v.ResolveRevision(head)
	if err == nil && headid != "" {
		if common.IsHead(head) {
			branch = strings.TrimPrefix(head, config.RefsHeads)
		}
	}

	// We have a branch, check whether tracking branch is set properly.
	if branch != "" {
		track = v.TrackBranch(branch)
	}

	log.Debugf("%sfetching (head: %s, branch: %s, track: %s, headid: %s, revid: %s, revision: %s)",
		v.Prompt(), head, branch, track, headid, revid, v.Revision)

	PostUpdate := func(update bool) error {
		var (
			err    error
			remote = v.GetDefaultRemote(true)
		)

		if branch != "" && track != v.Revision {
			if v.Revision != "" {
				if track == "" {
					log.Notef("%smanifest switched to %s", v.Prompt(), v.Revision)
				} else {
					log.Notef("%smanifest switched %s...%s", v.Prompt(), track, v.Revision)
				}
			} else {
				log.Notef("%smanifest no longer tracks %s", v.Prompt(), track)
			}

			log.Debugf("%supdating tracking of %s to %s/%s", v.Prompt(), branch, v.RemoteName, v.Revision)
			// Update remote tracking, or delete tracking if v.Revision is empty
			v.UpdateBranchTracking(branch, v.RemoteName, v.Revision)
		}

		if update && o.Submodules {
			err = v.SubmoduleUpdate()
			if err != nil {
				return err
			}
		}

		// Install gerrit hooks
		if remote != nil {
			if remote.GetType() == helper.ProtoTypeGerrit {
				v.InstallGerritHooks()
			}

			// Disable default push, push command must have specific refspec
			if remote.GetType() != "" {
				v.DisableDefaultPush()
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
			if !o.DetachHead {
				return PostUpdate(false)
			}
		}

		if headid != "" {
			localChanges, err := v.Revlist(headid, "--not", revid)
			if err != nil {
				log.Warnf("%srev-list failed: %s", v.Prompt(), err)
			}
			if len(localChanges) > 0 {
				log.Notef("%sdiscarding %d commits", v.Prompt(), len(localChanges))
			}
		}

		log.Debugf("%sdetached head, force checkout: %s", v.Prompt(), revid)
		err = v.CheckoutRevision(revid)
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
		log.Notef("%sleaving %s; does not track upstream", v.Prompt(), branch)
		err = v.CheckoutRevision(revid)
		if err != nil {
			return err
		}
		return PostUpdate(true)
	}

	log.Debugf("%schecking rev-list: %s..%s", v.Prompt(), headid, revid)
	remoteChanges, err := v.Revlist(revid, "--not", headid)
	if err != nil {
		log.Errorf("%srev-list failed: %s", v.Prompt(), err)
	}

	if !o.IsManifest || v.Revision == track {
		// No remote changes, no update.
		if len(remoteChanges) == 0 {
			log.Debugf("%sno remote changes found for project %s", v.Prompt(), v.Name)
			return PostUpdate(false)
		}
	}

	// Manifest project do not need to check publish ref.
	if !o.IsManifest {
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
				log.Debugf("%shas %d unpublished commit(s) for project %s",
					v.Prompt(),
					len(notMerged),
					v.Name)
				if len(remoteChanges) > 0 {
					log.Errorf("%sbranch %s is published (but not merged) and is now "+
						"%d commits behind",
						v.Prompt(),
						branch,
						len(remoteChanges))
				}
				return fmt.Errorf("branch %s is published (but not merged)", branch)
			}
			// Since last published, no other local changes.
			if pubid == headid {
				log.Debugf("%sall local commits are published", v.Prompt())
				err = v.FastForward(revid)
				if err != nil {
					return err
				}
				return PostUpdate(true)
			}
		}
	}

	// Failed if worktree is dirty.
	if !v.IsClean() {
		return fmt.Errorf("worktree of %s is dirty, checkout failed", v.Name)
	}

	// For ManifestProject, use `reset --hard` to switch branch,
	// no need to resolve conflict on a manifest project.
	if o.IsManifest && v.Revision != track {
		trackid, err := v.ResolveRemoteTracking(track)
		localChanges, err := v.Revlist(headid, "--not", trackid)
		if len(localChanges) > 0 {
			return fmt.Errorf("add --detach option to `git repo init` to throw away changes in '.repo/manifests'")
		}
		err = v.HardReset(revid)
		if err != nil {
			return err
		}
		return PostUpdate(false)
	}

	// Default action if not turn off by rebase attribute of project in manifest file.
	if v.IsRebase() {
		err = v.Rebase(revid)
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

// InstallGerritHooks installs gerrit hooks if remote of current project is gerrit.
func (v Project) InstallGerritHooks() error {
	hooksDir, err := config.GetRepoHooksDir()
	if err != nil {
		return err
	}
	if !path.Exist(hooksDir) {
		log.Warnf("cannot find hooks in %s", hooksDir)
		return nil
	}

	// Hooks dir of work repository maybe a symlink to object repository
	localHooksDir := filepath.Join(v.CommonDir(), "hooks")
	if p, err := filepath.EvalSymlinks(localHooksDir); err == nil {
		localHooksDir = p
	}
	log.Debugf("installing gerrit hooks for %s", v.Path)
	for name := range config.GerritHooks {
		src := filepath.Join(hooksDir, name)
		target := filepath.Join(localHooksDir, name)
		srcRel, err := filepath.Rel(localHooksDir, src)
		if err != nil {
			srcRel = src
		}
		if path.Exist(target) {
			linkedSrc, err := os.Readlink(target)
			if err != nil || filepath.Base(linkedSrc) != name {
				log.Debugf("%swill remove %s before recreate link",
					v.Prompt(),
					target)
				os.Remove(target)
			} else {
				continue
			}
		}
		err = os.Symlink(srcRel, target)
		if err != nil {
			return err
		}
	}
	return nil
}

// CopyFile copy files from src to dest.
func (v Project) CopyFile(src, dest string) error {
	srcAbs := filepath.Clean(filepath.Join(v.WorkDir, src))
	destAbs := filepath.Clean(filepath.Join(v.TopDir(), dest))

	if !strings.HasPrefix(srcAbs, v.TopDir()) {
		return fmt.Errorf("fail to copy file, src file '%s' beyond repo root '%s'", src, v.TopDir())
	}

	if !strings.HasPrefix(destAbs, v.TopDir()) {
		return fmt.Errorf("fail to copy file, dest file '%s' beyond repo root '%s'", dest, v.TopDir())
	}

	finfo, err := os.Stat(srcAbs)
	if err != nil {
		return nil
	}

	if !path.Exist(filepath.Dir(destAbs)) {
		os.MkdirAll(filepath.Dir(destAbs), 0755)
	}

	srcFile, err := os.Open(srcAbs)
	if err != nil {
		return fmt.Errorf("fail to open '%s': %s", srcAbs, err)
	}
	defer srcFile.Close()

	destFile, err := file.New(destAbs).SetPerm(finfo.Mode()).OpenCreateRewrite()
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

// LinkFile copy files from src to dest.
func (v Project) LinkFile(src, dest string) error {
	srcAbs := filepath.Clean(filepath.Join(v.WorkDir, src))
	destAbs := filepath.Clean(filepath.Join(v.TopDir(), dest))

	if !strings.HasPrefix(srcAbs, v.TopDir()) {
		return fmt.Errorf("fail to copy file, src file '%s' beyond repo root '%s'", src, v.TopDir())
	}

	if !strings.HasPrefix(destAbs, v.TopDir()) {
		return fmt.Errorf("fail to copy file, dest file '%s' beyond repo root '%s'", dest, v.TopDir())
	}

	_, err := os.Stat(srcAbs)
	if err != nil {
		return nil
	}

	destDir := filepath.Dir(destAbs)
	if !path.Exist(destDir) {
		os.MkdirAll(destDir, 0755)
	}
	srcRel, err := filepath.Rel(destDir, srcAbs)
	if err != nil {
		srcRel = srcAbs
	}
	if path.Exist(destAbs) {
		os.Remove(destAbs)
	}
	if cap.CanSymlink() {
		return os.Symlink(srcRel, destAbs)
	}
	return os.Link(srcRel, destAbs)
}

// CopyAndLinkFiles copies and links files.
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
