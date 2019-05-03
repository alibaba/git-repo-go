package project

import (
	"fmt"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"github.com/jiangxin/multi-log"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Branch wraps branch name and object ID
type Branch struct {
	Name string
	Hash string
}

// ShortName removes prefix "refs/heads/"
func (v Branch) ShortName() string {
	return strings.TrimPrefix(v.Name, config.RefsHeads)
}

// Reference wraps reference name and object ID
type Reference struct {
	Name string
	Hash string
}

// ReviewableBranch holds branch of proect ready for upload
type ReviewableBranch struct {
	Project  *Project
	Branch   Branch
	Track    Reference
	Uploaded bool
	Error    error
}

// AppendReviewers adds reviewers to people
func (v ReviewableBranch) AppendReviewers(people [][]string) {
	cfg := v.Project.Config()
	review := v.Project.Remote.GetRemote().Review

	key := fmt.Sprintf("review.%s.autoreviewer", review)
	reviewers := cfg.Get(key)
	if reviewers != "" {
		for _, reviewer := range strings.Split(reviewers, ",") {
			reviewer = strings.TrimSpace(reviewer)
			people[0] = append(people[0], reviewer)
		}
	}

	key = fmt.Sprintf("review.%s.autocopy", review)
	reviewers = cfg.Get(key)
	if reviewers != "" {
		for _, reviewer := range strings.Split(reviewers, ",") {
			reviewer = strings.TrimSpace(reviewer)
			people[1] = append(people[1], reviewer)
		}
	}
}

// Published returns published reference
func (v ReviewableBranch) Published() *Reference {
	pub := Reference{}
	pub.Name = config.RefsPub + v.Branch.ShortName()
	revid, err := v.Project.ResolveRevision(pub.Name)
	if err != nil {
		return nil
	}

	pub.Hash = revid
	return &pub
}

// Commits contains commits avaiable for review
func (v ReviewableBranch) Commits() []string {
	commits, err := v.Project.Revlist(v.Branch.Hash, "--not", v.Track.Hash)
	if err != nil {
		log.Errorf("fail to get commits of ReviewableBranch %s: %s", v.Branch, err)
		return nil
	}
	return commits
}

// Heads returns branches of repository
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

// RemoteTrackBranch gets remote tracking branch
func (v Repository) RemoteTrackBranch(branch string) string {
	if branch == "" {
		branch = v.GetHead()
	}
	if branch == "" {
		return ""
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	cfg := v.Config()
	return cfg.Get("branch." + branch + ".merge")
}

// LocalTrackRemoteBranch gets local tracking remote branch
func (v Repository) LocalTrackRemoteBranch(branch string) string {
	if branch == "" {
		branch = v.GetHead()
	}
	if branch == "" {
		return ""
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	cfg := v.Config()
	track := cfg.Get("branch." + branch + ".merge")
	track = strings.TrimPrefix(track, config.RefsHeads)
	remote := cfg.Get("branch." + branch + ".remote")
	return config.RefsRemotes + remote + "/" + track
}

// DeleteBranch deletes a branch
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

// GetHead returns head branch.
func (v Project) GetHead() string {
	return v.WorkRepository.GetHead()
}

// Heads returns branches of repository
func (v Project) Heads() []Branch {
	return v.WorkRepository.Heads()
}

// RemoteTracking returns name of current remote tracking branch
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

// DeleteBranch deletes a branch
func (v Project) DeleteBranch(branch string) error {
	return v.WorkRepository.DeleteBranch(branch)
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

// GetUploadableBranch returns branch which has commits ready for upload
func (v *Project) GetUploadableBranch(branch string) *ReviewableBranch {
	if branch == "" {
		branch = v.GetHead()
		if branch == "" {
			return nil
		}
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)
	remote := v.Config().Get("branch." + branch + ".remote")

	if v.Remote == nil || v.Remote.GetType() == config.RemoteTypeUnknown {
		log.Debugf("unknown type of remote '%s' for project '%s'",
			remote,
			v.Path)
		return nil
	}

	manifestRemote := v.Remote.GetRemote().Name

	if remote != manifestRemote {
		log.Warnf("unmatch remote: remote of branch '%s' is '%s', not '%s'",
			branch,
			remote,
			manifestRemote,
		)
		return nil
	}

	branchID, err := v.ResolveRevision(branch)
	if err != nil {
		return nil
	}
	track := v.LocalTrackRemoteBranch(branch)
	if track == "" {
		return nil
	}
	trackID, err := v.ResolveRevision(track)
	if err != nil {
		return nil
	}

	rb := ReviewableBranch{
		Project: v,
		Branch: Branch{
			Name: branch,
			Hash: branchID},
		Track: Reference{
			Name: track,
			Hash: trackID},
	}

	if len(rb.Commits()) == 0 {
		return nil
	}

	pub := rb.Published()
	if pub != nil && pub.Hash == branchID {
		return nil
	}

	return &rb
}

// GetUploadableBranches returns branches which has commits ready for upload
func (v *Project) GetUploadableBranches(branch string) []ReviewableBranch {
	var (
		avail = []ReviewableBranch{}
	)

	if branch != "" {
		rb := v.GetUploadableBranch(branch)
		if rb == nil {
			return nil
		}
		avail = append(avail, *rb)
		return avail
	}

	for _, head := range v.Heads() {
		rb := v.GetUploadableBranch(head.Name)
		if rb == nil {
			continue
		}
		avail = append(avail, *rb)
	}

	return avail
}

// RemoteTrackBranch gets remote tracking branch
func (v Project) RemoteTrackBranch(branch string) string {
	return v.WorkRepository.RemoteTrackBranch(branch)
}

// LocalTrackRemoteBranch gets local tracking remote branch
func (v Project) LocalTrackRemoteBranch(branch string) string {
	return v.WorkRepository.LocalTrackRemoteBranch(branch)
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
