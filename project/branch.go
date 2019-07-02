package project

import (
	"fmt"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	log "github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Branch wraps branch name and object ID.
type Branch struct {
	Name string
	Hash string
}

// ShortName removes prefix "refs/heads/".
func (v Branch) ShortName() string {
	return strings.TrimPrefix(v.Name, config.RefsHeads)
}

// Reference wraps reference name and object ID.
type Reference struct {
	Name string
	Hash string
}

// Heads returns branches of repository.
func (v Repository) Heads() []Branch {
	var heads []Branch
	branches, err := v.Raw().Branches()
	if err != nil {
		return nil
	}

	branches.ForEach(func(branch *plumbing.Reference) error {
		head := Branch{
			Name: string(branch.Name()),
			Hash: branch.Hash().String(),
		}
		heads = append(heads, head)
		return nil
	})
	return heads
}

// TrackBranch gets remote tracking branch.
func (v Repository) TrackBranch(branch string) string {
	if branch == "" {
		branch = v.GetHead()
	}
	if branch == "" {
		return ""
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	cfg := v.Config()
	return strings.TrimPrefix(cfg.Get("branch."+branch+".merge"), config.RefsHeads)
}

// TrackRemote gets the remote name what current branch is tracking.
func (v Repository) TrackRemote(branch string) string {
	if branch == "" {
		branch = v.GetHead()
	}
	if branch == "" {
		return ""
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	cfg := v.Config()
	return cfg.Get("branch." + branch + ".remote")
}

// LocalTrackBranch gets local tracking remote branch
func (v Repository) LocalTrackBranch(branch string) string {
	if branch == "" {
		branch = v.GetHead()
	}
	if branch == "" {
		return ""
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	cfg := v.Config()
	track := strings.TrimPrefix(cfg.Get("branch."+branch+".merge"), config.RefsHeads)
	remote := cfg.Get("branch." + branch + ".remote")
	if remote == "" || track == "" {
		return ""
	}
	return config.RefsRemotes + remote + "/" + track
}

// DeleteBranch deletes a branch.
func (v Repository) DeleteBranch(branch string) error {
	// TODO: go-git fail to delete a branch
	// TODO: return v.Raw().DeleteBranch(branch)
	if IsHead(branch) {
		branch = strings.TrimPrefix(branch, config.RefsHeads)
	}
	cmdArgs := []string{
		GIT,
		"branch",
		"-D",
		branch,
		"--",
	}
	return executeCommandIn(v.Path, cmdArgs)
}

// UpdateRef creates new reference.
func (v Repository) UpdateRef(refname, base, reason string) error {
	var (
		err error
		ref *plumbing.Reference
	)

	if config.IsDryRun() {
		log.Notef("will update-ref %s on %s, reason: %s", refname, base, reason)
		return nil
	}

	raw := v.Raw()

	if IsSha(base) {
		ref = plumbing.NewHashReference(plumbing.ReferenceName(refname),
			plumbing.NewHash(base))
	} else {
		hash, err := raw.ResolveRevision(plumbing.Revision(base))
		if err != nil {
			return fmt.Errorf("cannot resolve base rev '%s' when update ref: %s",
				base,
				err)
		}
		ref = plumbing.NewHashReference(plumbing.ReferenceName(refname), *hash)
	}
	err = raw.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("fail create '%s' on '%s': %s", refname, base, err)
	}

	return nil
}

// GetHead returns head branch.
func (v Project) GetHead() string {
	return v.Head()
}

// Heads returns branches of repository.
func (v Project) Heads() []Branch {
	return v.WorkRepository.Heads()
}

// RemoteTracking returns name of current remote tracking branch.
func (v Project) RemoteTracking(rev string) string {
	if rev == "" || IsSha(rev) {
		return ""
	}
	if IsHead(rev) {
		rev = strings.TrimPrefix(rev, config.RefsHeads)
	}
	if IsRef(rev) {
		return ""
	}
	return v.Config().Get("branch." + rev + ".merge")
}

// ResolveRevision checks and resolves reference to revid.
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

// ResolveRemoteTracking returns revision id of current remote tracking branch.
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
				v.RemoteName,
				rev)
		}
	}
	revid, err := raw.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", fmt.Errorf("revision %s in %s not found", rev, v.Name)
	}
	return revid.String(), nil
}

// DeleteBranch deletes a branch.
func (v Project) DeleteBranch(branch string) error {
	return v.WorkRepository.DeleteBranch(branch)
}

// UpdateRef creates new reference.
func (v Project) UpdateRef(ref, base, reason string) error {
	return v.WorkRepository.UpdateRef(ref, base, reason)
}

// StartBranch creates new branch.
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
	remote := v.RemoteName
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
	v.UpdateBranchTracking(branch, remote, track)
	return nil
}

// TrackBranch gets remote track branch name.
func (v Project) TrackBranch(branch string) string {
	return v.WorkRepository.TrackBranch(branch)
}

// LocalTrackBranch gets local tracking remote branch.
func (v Project) LocalTrackBranch(branch string) string {
	return v.WorkRepository.LocalTrackBranch(branch)
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
