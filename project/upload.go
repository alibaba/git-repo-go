package project

import (
	"fmt"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"github.com/jiangxin/multi-log"
)

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
