package project

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	log "github.com/jiangxin/multi-log"
)

// UploadOptions is options for upload related methods.
type UploadOptions struct {
	AutoTopic    bool
	Description  string
	DestBranch   string
	Draft        bool
	Issue        string
	MockGitPush  bool
	NoCertChecks bool
	NoEmails     bool
	People       [][]string
	Private      bool
	PushOptions  []string
	Title        string
	UserEmail    string
	WIP          bool
}

// RemoteTrack holds info of remote tracking branch
type RemoteTrack struct {
	Remote string
	Branch string
	Track  Reference
}

// ReviewableBranch holds branch of project ready for upload.
type ReviewableBranch struct {
	Project     *Project
	Branch      Branch
	DestBranch  string
	RemoteTrack RemoteTrack
	Uploaded    bool
	Error       error

	isPublished int
}

// IsPublished indicates a branch has been published.
func (v *ReviewableBranch) IsPublished() bool {
	if v.isPublished == 0 {
		ref := v.Published()
		if ref != nil {
			v.isPublished = 1
		} else {
			v.isPublished = -1
		}
	}
	if v.isPublished == 1 {
		return true
	}
	return false
}

// AppendReviewers adds reviewers to people.
func (v ReviewableBranch) AppendReviewers(people [][]string) {
	cfg := v.Project.ConfigWithDefault()
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

// Published returns published reference.
func (v *ReviewableBranch) Published() *Reference {
	pub := Reference{}
	pub.Name = config.RefsPub + v.Branch.ShortName()
	revid, err := v.Project.ResolveRevision(pub.Name)
	if err != nil {
		v.isPublished = -1
		return nil
	}

	v.isPublished = 1
	pub.Hash = revid
	return &pub
}

// Commits contains commits avaiable for review.
func (v ReviewableBranch) Commits() []string {
	commits, err := v.Project.Revlist(v.Branch.Hash, "--not", v.RemoteTrack.Track.Hash)
	if err != nil {
		log.Errorf("fail to get commits of ReviewableBranch %s: %s", v.Branch, err)
		return nil
	}
	return commits
}

// UploadForReview sends review for branch.
func (v ReviewableBranch) UploadForReview(o *UploadOptions, people [][]string) error {
	var err error

	p := v.Project
	if p == nil {
		return fmt.Errorf("no project for reviewable branch")
	}
	if p.Remote == nil {
		return fmt.Errorf("no remote for project '%s' for review", p.Name)
	}
	manifestRemote := p.Remote.GetRemote()
	if manifestRemote.Review == "" {
		return fmt.Errorf("project '%s' has no review url", p.Name)
	}

	if o.DestBranch == "" {
		o.DestBranch = v.DestBranch
		if o.DestBranch == "" {
			return fmt.Errorf("no destination for review")
		}
	}
	o.People = people

	cmdArgs, err := p.Remote.UploadCommands(o, &v)
	if err != nil {
		return err
	}

	if config.IsDryRun() || o.MockGitPush {
		log.Notef("will execute command: %s", strings.Join(cmdArgs, " "))
	} else {
		log.Debugf("review by command: %s", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = p.WorkDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("upload failed: %s", err)
		}
	}

	branchName := v.Branch.Name
	if strings.HasPrefix(branchName, config.RefsHeads) {
		branchName = strings.TrimPrefix(branchName, config.RefsHeads)
	}
	msg := fmt.Sprintf("review from %s to %s on %s",
		branchName,
		o.DestBranch,
		manifestRemote.Review)

	log.Debugf("Update reference '%s': %s", config.RefsPub+branchName, msg)
	err = p.UpdateRef(config.RefsPub+branchName,
		config.RefsHeads+branchName,
		msg)

	if err != nil {
		return fmt.Errorf("fail to create reference '%s': %s",
			config.RefsPub+branchName,
			err)
	}
	return nil
}

// GetUploadableBranch returns branch which has commits ready for upload.
func (v *Project) GetUploadableBranch(branch, remote, remoteBranch string) *ReviewableBranch {
	if branch == "" {
		branch = v.GetHead()
		if branch == "" {
			return nil
		}
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)
	if remote == "" {
		remote = v.Config().Get("branch." + branch + ".remote")
	}
	if remoteBranch == "" {
		remoteBranch = v.Config().Get("branch." + branch + ".merge")
	}

	manifestRemote := ""
	if v.Remote != nil {
		manifestRemote = v.Remote.GetRemote().Name
		if remote != manifestRemote && !config.IsSingleMode() {
			log.Warnf("cannot upload, unmatch remote for '%s': %s != %s",
				branch,
				remote,
				manifestRemote,
			)
			return nil
		}
	}

	branchID, err := v.ResolveRevision(branch)
	if err != nil {
		return nil
	}
	track := v.RemoteMatchingBranch(remote, remoteBranch)
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
		DestBranch: v.TrackBranch(branch),
		RemoteTrack: RemoteTrack{
			Remote: remote,
			Branch: remoteBranch,
			Track: Reference{
				Name: track,
				Hash: trackID,
			},
		},
	}

	if len(rb.Commits()) == 0 {
		return nil
	}

	pub := rb.Published()
	if pub != nil && pub.Hash == branchID {
		log.Notef("no change in project %s (branch %s) since last upload",
			v.Path,
			branch)
		return nil
	}

	return &rb
}

// GetUploadableBranches returns branches which has commits ready for upload.
func (v *Project) GetUploadableBranches(branch string) []ReviewableBranch {
	var (
		avail = []ReviewableBranch{}
	)

	if branch != "" {
		rb := v.GetUploadableBranch(branch, "", "")
		if rb == nil {
			return nil
		}
		avail = append(avail, *rb)
		return avail
	}

	for _, head := range v.Heads() {
		rb := v.GetUploadableBranch(head.Name, "", "")
		if rb == nil {
			continue
		}
		avail = append(avail, *rb)
	}

	return avail
}
