package project

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Project inherits manifest's Project and has two related repositories
type Project struct {
	manifest.Project

	WorkDir          string
	ObjectRepository *Repository
	WorkRepository   *Repository
	Settings         *RepoSettings
}

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

// CheckoutOptions is options for git fetch
type CheckoutOptions struct {
	RepoSettings

	Quiet      bool
	DetachHead bool
}

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

// RepoRoot returns root dir of repo workspace.
func (v Project) RepoRoot() string {
	return v.Settings.RepoRoot
}

// ManifestURL returns manifest URL
func (v Project) ManifestURL() string {
	return v.Settings.ManifestURL
}

// ReferencePath returns path of reference git dir
func (v Project) ReferencePath() string {
	var (
		rdir = ""
		err  error
	)

	if v.Settings.Reference == "" {
		return ""
	}

	if !filepath.IsAbs(v.Settings.Reference) {
		v.Settings.Reference, err = path.Abs(v.Settings.Reference)
		if err != nil {
			log.Errorf("bad reference path '%s': %s", v.Settings.Reference, err)
			v.Settings.Reference = ""
			return ""
		}
	}

	if !v.IsMetaProject() {
		rdir = filepath.Join(v.Settings.Reference, v.Name+".git")
		if path.Exist(rdir) {
			return rdir
		}
		rdir = filepath.Join(v.Settings.Reference,
			config.DotRepo,
			config.Projects,
			v.Path+".git")
		if path.Exist(rdir) {
			return rdir
		}
		return ""
	}

	if v.Settings.ManifestURL != "" {
		u, err := url.Parse(v.Settings.ManifestURL)
		if err == nil {
			dir := u.RequestURI()
			if !strings.HasSuffix(dir, ".git") {
				dir += ".git"
			}
			dirs := strings.Split(dir, "/")
			for i := 1; i < len(dirs); i++ {
				dir = filepath.Join(v.Settings.Reference, filepath.Join(dirs[i:]...))
				if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
					rdir = dir
					break
				}
			}
		}
	}

	if rdir == "" {
		dir := filepath.Join(v.Settings.Reference, config.DotRepo, config.ManifestsDotGit)
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			rdir = dir
		}
	}

	return rdir
}

// Exists indicates whether project exists or not.
func (v Project) Exists() bool {
	if _, err := os.Stat(v.WorkDir); err != nil {
		return false
	}

	if _, err := os.Stat(filepath.Join(v.WorkDir, ".git")); err != nil {
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
			v.WorkRepository.Init(v.Remote, remoteURL, referenceGitDir)
		} else {
			v.WorkRepository.InitByLink(v.Remote, remoteURL, v.ObjectRepository)
		}
	}

	// TODO: install hooks
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
		err = v.WorkRepository.Fetch(v.Remote, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrepareWorkdir setup workdir layout: .git is gitdir file points to repository
func (v *Project) PrepareWorkdir() error {
	err := os.MkdirAll(v.WorkDir, 0755)
	if err != nil {
		return nil
	}

	gitdir := filepath.Join(v.WorkDir, ".git")
	if _, err = os.Stat(gitdir); err != nil {
		// Remove index file for fresh checkout
		idxfile := filepath.Join(v.WorkRepository.Path, "index")
		err = os.Remove(idxfile)

		relDir, err := filepath.Rel(v.WorkDir, v.WorkRepository.Path)
		if err != nil {
			relDir = v.WorkRepository.Path
		}
		err = ioutil.WriteFile(gitdir,
			[]byte("gitdir: "+relDir+"\n"),
			0644)
		if err != nil {
			return fmt.Errorf("fail to create gitdir for %s: %s", v.Name, err)
		}
	}
	return nil
}

// CleanPublishedCache removes obsolete refs/published/ references.
func (v Project) CleanPublishedCache() error {
	var err error

	raw := v.WorkRepository.Raw()
	if raw == nil {
		return nil
	}
	pubMap := make(map[string]string)
	headsMap := make(map[string]string)
	refs, _ := raw.References()
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			if strings.HasPrefix(string(ref.Name()), config.RefsPub) {
				pubMap[string(ref.Name())] = ref.Hash().String()
			} else if strings.HasPrefix(string(ref.Name()), config.RefsHeads) {
				headsMap[string(ref.Name())] = ref.Hash().String()
			}
			fmt.Println(ref)
		}
		return nil
	})

	for name := range pubMap {
		branch := config.RefsHeads + strings.TrimPrefix(name, config.RefsPub)
		if _, ok := headsMap[branch]; !ok {
			log.Infof("will delete obsolete ref: %s", name)
			err = raw.Storer.RemoveReference(plumbing.ReferenceName(name))
			if err != nil {
				log.Errorf("fail to remove reference '%s'", name)
			} else {
				log.Infof("removed reference '%s'", name)
			}
		}
	}

	return nil
}

// ResolveRevision checks and resolves reference to revid
func (v Project) ResolveRevision(rev string) (string, error) {
	if rev == "" {
		return "", nil
	}

	raw := v.WorkRepository.Raw()
	if raw == nil {
		return "", fmt.Errorf("repository for %s is missing, fail to parse %s", v.Name, rev)
	}

	if rev == "" {
		log.Errorf("empty revision to resolve for proejct '%s'", v.Name)
	}

	revid, err := raw.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}
	return revid.String(), nil
}

// ResolveRemoteTracking returns revision id of current remote tracking branch
func (v Project) ResolveRemoteTracking(rev string) (string, error) {
	raw := v.WorkRepository.Raw()
	if raw == nil {
		return "", fmt.Errorf("repository for %s is missing, fail to parse %s", v.Name, v.Revision)
	}

	if rev == "" {
		log.Errorf("empty Revision for proejct '%s'", v.Name)
	}
	if !IsSha(rev) {
		if IsHead(rev) {
			rev = strings.TrimPrefix(rev, config.RefsHeads)
		}
		if !strings.HasPrefix(rev, config.Refs) {
			rev = fmt.Sprintf("%s%s/%s",
				config.RefsRemotes,
				v.Remote,
				rev)
		}
	}
	revid, err := raw.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", fmt.Errorf("revision %s in %s not found", rev, v.Name)
	}
	return revid.String(), nil
}

// GetHead returns head branch.
func (v Project) GetHead() string {
	return v.WorkRepository.GetHead()
}

// IsRebaseInProgress checks whether is in middle of a rebase.
func (v Project) IsRebaseInProgress() bool {
	return v.WorkRepository.IsRebaseInProgress()
}

// RevisionIsValid checks if revision is valid
func (v Project) RevisionIsValid(revision string) bool {
	return v.WorkRepository.RevisionIsValid(revision)
}

// Revlist works like rev-list
func (v Project) Revlist(args ...string) ([]string, error) {
	return v.WorkRepository.Revlist(args...)
}

// RemoteTrackBranch gets remote tracking branch
func (v Project) RemoteTrackBranch(branch string) string {
	return v.WorkRepository.RemoteTrackBranch(branch)
}

// PublishedReference forms published reference for specific branch.
func (v Project) PublishedReference(branch string) string {
	pub := config.RefsPub + branch

	if v.RevisionIsValid(pub) {
		return pub
	}
	return ""
}

// PublishedRevision resolves published reference to revision id.
func (v Project) PublishedRevision(branch string) string {
	raw := v.WorkRepository.Raw()
	pub := config.RefsPub + branch

	if raw == nil {
		return ""
	}

	revid, err := raw.ResolveRevision(plumbing.Revision(pub))
	if err == nil {
		return revid.String()
	}
	return ""
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

// UpdateBranchTracking updates branch tracking info.
func (v Project) UpdateBranchTracking(branch, remote, track string) {
	cfg := v.Config()
	if track == "" ||
		IsSha(track) ||
		(IsRef(track) && !IsHead(track)) {
		cfg.Unset("branch." + branch + ".merge")
		cfg.Unset("branch." + branch + ".remote")
		v.SaveConfig(cfg)
		return
	}

	if !IsHead(track) {
		track = config.RefsHeads + track
	}

	cfg.Set("branch."+branch+".merge", track)
	if remote != "" {
		cfg.Set("branch."+branch+".remote", remote)
	}

	v.SaveConfig(cfg)
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
	if cap.Symlink() {
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
			v.UpdateBranchTracking(branch, v.Remote, v.Revision)
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

// GitRepository returns go-git's repository object for project worktree
func (v *Project) GitRepository() (*git.Repository, error) {
	return git.PlainOpen(v.WorkDir)
}

// GitWorktree returns go-git's worktree oject
func (v *Project) GitWorktree() (*git.Worktree, error) {
	r, err := v.GitRepository()
	if err != nil {
		return nil, err
	}
	return r.Worktree()
}

// Head returns current branch of project's workdir
func (v *Project) Head() string {
	r, err := v.GitRepository()
	if err != nil {
		return ""
	}

	// Not checkout yet
	head, err := r.Head()
	if head == nil {
		return ""
	}
	return head.Name().String()
}

// SetManifestURL sets manifestURL and change remote url if is MetaProject
func (v *Project) SetManifestURL(manifestURL string) error {
	if manifestURL != "" && !strings.HasSuffix(manifestURL, ".git") {
		manifestURL += ".git"
	}
	if v.Settings.ManifestURL == manifestURL || manifestURL == "" {
		return nil
	}
	v.Settings.ManifestURL = manifestURL
	if v.IsMetaProject() {
		return v.SetGitRemoteURL(manifestURL)
	}
	return nil
}

// SetGitRemoteURL sets remote.<remote>.url setting in git config
func (v *Project) SetGitRemoteURL(remoteURL string) error {
	remote := v.Remote
	if remote == "" {
		remote = "origin"
	}

	repo := v.WorkRepository
	if repo == nil {
		return fmt.Errorf("project '%s' has no working repository", v.Name)
	}

	cfg := repo.Config()
	cfg.Set("remote."+remote+".url", remoteURL)
	return repo.SaveConfig(cfg)
}

// GitConfigRemoteURL returns remote.<remote>.url setting in git config
func (v *Project) GitConfigRemoteURL() string {
	remote := v.Remote
	if remote == "" {
		remote = "origin"
	}

	repo := v.WorkRepository
	return repo.GitConfigRemoteURL(remote)
}

// GetRemoteURL returns new remtoe url user provided or from manifest repo url
func (v *Project) GetRemoteURL() (string, error) {
	if v.Settings.ManifestURL == "" && v.IsMetaProject() {
		v.Settings.ManifestURL = v.GitConfigRemoteURL()
	}
	if v.Settings.ManifestURL == "" {
		return "", fmt.Errorf("project '%s' has empty manifest url", v.Name)
	}
	if v.IsMetaProject() {
		return v.Settings.ManifestURL, nil
	}
	if v.GetRemote() == nil {
		return "", fmt.Errorf("project '%s' has no remote '%s'", v.Name, v.Remote)
	}

	u, err := urlJoin(v.Settings.ManifestURL, v.GetRemote().Fetch, v.Name+".git")
	if err != nil {
		return "", fmt.Errorf("fail to remote url for '%s': %s", v.Name, err)
	}
	return u, nil
}

// Config returns git config file parser
func (v *Project) Config() goconfig.GitConfig {
	return v.WorkRepository.Config()
}

// SaveConfig will save config to git config file
func (v *Project) SaveConfig(cfg goconfig.GitConfig) error {
	return v.WorkRepository.SaveConfig(cfg)
}

// MatchGroups indecates if project belongs to special groups
func (v Project) MatchGroups(expect string) bool {
	return MatchGroups(expect, v.Groups)
}

// GetSubmoduleProjects returns submodule projects
func (v Project) GetSubmoduleProjects() []*Project {
	// TODO: return submodule projects
	log.Panic("not implement GitSubmodules")
	return nil
}

// NewProject returns a project: project worktree with a bared repo and a seperate repository
func NewProject(project *manifest.Project, s *RepoSettings) *Project {
	var (
		objectRepoPath string
		workRepoPath   string
	)

	if s.ManifestURL != "" && !strings.HasSuffix(s.ManifestURL, ".git") {
		s.ManifestURL += ".git"
	}
	p := Project{
		Project:  *project,
		Settings: s,
	}

	if !p.IsMetaProject() && s.ManifestURL == "" {
		log.Panicf("unknown remote url for %s", p.Name)
	}

	if p.IsMetaProject() {
		p.WorkDir = filepath.Join(p.RepoRoot(), config.DotRepo, p.Path)
	} else {
		p.WorkDir = filepath.Join(p.RepoRoot(), p.Path)
	}

	if !p.IsMetaProject() && cap.Symlink() {
		objectRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			config.ProjectObjects,
			p.Name+".git",
		)
		p.ObjectRepository = &Repository{
			Name:     p.Name,
			RelPath:  p.Path,
			Path:     objectRepoPath,
			IsBare:   true,
			Settings: s,
		}
	}

	if p.IsMetaProject() {
		workRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			p.Path+".git",
		)
	} else {
		workRepoPath = filepath.Join(
			p.RepoRoot(),
			config.DotRepo,
			config.Projects,
			p.Path+".git",
		)
	}

	p.WorkRepository = &Repository{
		Name:      p.Name,
		RelPath:   p.Path,
		Path:      workRepoPath,
		Remote:    p.Remote,
		Revision:  p.Revision,
		IsBare:    false,
		Reference: p.ReferencePath(),
		Settings:  s,
	}

	remoteURL, err := p.GetRemoteURL()
	if err != nil {
		log.Panicf("fail to get remote url for '%s': %s", p.Name, err)
	}
	p.WorkRepository.RemoteURL = remoteURL

	return &p
}

// NewMirrorProject returns a mirror project
func NewMirrorProject(project *manifest.Project, s *RepoSettings) *Project {
	var (
		repoPath string
	)

	if s.ManifestURL != "" && !strings.HasSuffix(s.ManifestURL, ".git") {
		s.ManifestURL += ".git"
	}
	p := Project{
		Project:  *project,
		Settings: s,
	}

	if !p.IsMetaProject() && s.ManifestURL == "" {
		log.Panicf("unknown remote url for %s", p.Name)
	}

	p.WorkDir = ""

	repoPath = filepath.Join(
		p.RepoRoot(),
		p.Name+".git",
	)

	p.WorkRepository = &Repository{
		Name:      p.Name,
		RelPath:   p.Path,
		Path:      repoPath,
		Remote:    p.Remote,
		Revision:  p.Revision,
		IsBare:    true,
		Reference: p.ReferencePath(),
		Settings:  s,
	}

	remoteURL, err := p.GetRemoteURL()
	if err != nil {
		log.Panicf("fail to get remote url for '%s': %s", p.Name, err)
	}
	p.WorkRepository.RemoteURL = remoteURL

	return &p
}

func isHashRevision(rev string) bool {
	re := regexp.MustCompile(`^[0-9][a-f]{7,}$`)
	return re.MatchString(rev)
}

// Join two group of projects, ignore duplicated projects
func Join(group1, group2 []*Project) []*Project {
	projectMap := make(map[string]bool)
	result := make([]*Project, len(group1)+len(group2))

	for _, p := range group1 {
		if _, ok := projectMap[p.Path]; !ok {
			projectMap[p.Path] = true
			result = append(result, p)
		}
	}
	for _, p := range group2 {
		if _, ok := projectMap[p.Path]; !ok {
			projectMap[p.Path] = true
			result = append(result, p)
		}
	}
	return result
}

// IndexByName returns a map using project name as index to group projects
func IndexByName(projects []*Project) map[string][]*Project {
	result := make(map[string][]*Project)
	for _, p := range projects {
		if _, ok := result[p.Name]; !ok {
			result[p.Name] = []*Project{p}
		} else {
			result[p.Name] = append(result[p.Name], p)
		}
	}
	return result
}

// IndexByPath returns a map using project path as index to group projects
func IndexByPath(projects []*Project) map[string]*Project {
	result := make(map[string]*Project)
	for _, p := range projects {
		result[p.Path] = p
	}
	return result
}

// Tree is used to group projects by path
type Tree struct {
	Path    string
	Project *Project
	Trees   []*Tree
}

// ProjectsTree returns a map using project path as index to group projects
func ProjectsTree(projects []*Project) *Tree {
	pMap := make(map[string]*Project)
	paths := []string{}
	for _, p := range projects {
		pMap[p.Path] = p
		paths = append(paths, p.Path)
	}
	sort.Strings(paths)

	root := &Tree{Path: "/"}
	treeAppend(root, paths, pMap)
	return root
}

func treeAppend(tree *Tree, paths []string, pMap map[string]*Project) {
	var (
		oldEntry, newEntry *Tree
		i, j               int
	)

	for i = 0; i < len(paths); i++ {
		for strings.HasSuffix(paths[i], "/") {
			paths[i] = strings.TrimSuffix(paths[i], "/")
		}
		newEntry = &Tree{
			Path:    paths[i],
			Project: pMap[paths[i]],
		}
		if i > 0 && strings.HasPrefix(paths[i], paths[i-1]+"/") {
			for j = i + 1; j < len(paths); j++ {
				if !strings.HasPrefix(paths[j], paths[i-1]+"/") {
					break

				}
			}
			treeAppend(oldEntry, paths[i:j], pMap)
			i = j - 1
			continue
		}
		oldEntry = newEntry
		tree.Trees = append(tree.Trees, newEntry)
	}
}

// StartBranch creates new branch
func (v Project) StartBranch(branch, track string) error {
	var err error

	if track == "" {
		track = v.Revision
	}
	if IsHead(branch) {
		branch = strings.TrimPrefix(branch, config.RefsHeads)
	}

	// Branch is already the current branch
	head := v.GetHead()
	if head == config.RefsHeads+branch {
		return nil
	}

	// Checkout if branch is already exist in repository
	if v.RevisionIsValid(config.RefsHeads + branch) {
		cmdArgs := []string{
			GIT,
			"checkout",
			branch,
			"--",
		}
		return executeCommandIn(v.WorkDir, cmdArgs)
	}

	// Get revid from already fetched tracking for v.Revision
	revid, err := v.ResolveRemoteTracking(v.Revision)
	remote := v.Remote
	if remote == "" {
		remote = "origin"
	}

	// Create a new branch
	cmdArgs := []string{
		GIT,
		"checkout",
		"-b",
		branch,
	}
	if revid != "" {
		cmdArgs = append(cmdArgs, revid)
	}
	cmdArgs = append(cmdArgs, "--")
	err = executeCommandIn(v.WorkDir, cmdArgs)
	if err != nil {
		return err
	}

	// Create remote tracking
	v.UpdateBranchTracking(branch, remote, v.Revision)
	return nil
}

// DetachHead makes detached HEAD
func (v Project) DetachHead() error {
	cmdArgs := []string{
		GIT,
		"checkout",
		"HEAD^0",
		"--",
	}
	return executeCommandIn(v.WorkDir, cmdArgs)
}

// DeleteBranch deletes a branch
func (v Project) DeleteBranch(branch string) error {
	return v.WorkRepository.DeleteBranch(branch)
}
