package project

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alibaba/git-repo-go/common"
	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/helper"
	log "github.com/jiangxin/multi-log"
)

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
	CodeReview  common.CodeReview // Push to update specific code review, only available for single repository mode.
	Remote      *Remote

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
	var (
		review string
	)

	cfg := v.Project.ConfigWithDefault()
	if v.Remote != nil {
		review = v.Remote.Review
	}
	if review == "" {
		return
	}

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
	if v.CodeReview.Empty() {
		pub.Name = config.RefsPub + v.Branch.ShortName()
	} else {
		pub.Name = v.CodeReview.Ref
	}
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
	var (
		commits []string
		err     error
	)

	if v.CodeReview.Empty() {
		commits, err = v.Project.Revlist(v.Branch.Hash, "--not", v.RemoteTrack.Track.Hash)
	} else {
		commits, err = v.Project.Revlist(v.Branch.Hash, "--not", v.CodeReview.Ref)
	}
	if err != nil {
		log.Errorf("%sfail to get commits of ReviewableBranch %s: %s",
			v.Project.Prompt(),
			v.Branch,
			err)
		return nil
	}
	return commits
}

// UploadForReview sends review for branch.
func (v ReviewableBranch) UploadForReview(o *common.UploadOptions) error {
	var err error

	p := v.Project
	if p == nil {
		return fmt.Errorf("no project for reviewable branch")
	}

	remoteName, remoteURL := p.GetRemotePushNameURL(v.Remote)
	if remoteURL == "" {
		return fmt.Errorf("project '%s' has no review url", p.Name)
	}
	gitURL := config.ParseGitURL(remoteURL)
	if gitURL == nil {
		return fmt.Errorf("bad review URL: %s", remoteURL)
	}
	o.RemoteName = remoteName
	o.RemoteURL = remoteURL

	if v.CodeReview.Empty() && o.DestBranch == "" {
		o.DestBranch = v.DestBranch
		if o.DestBranch == "" {
			return fmt.Errorf("no destination for review")
		}
	}

	pushCmd, err := v.Remote.GetGitPushCommand(o)
	if err != nil {
		return err
	}

	cmdArgs := []string{pushCmd.Cmd}
	if len(pushCmd.GitConfig) > 0 {
		for _, c := range pushCmd.GitConfig {
			cmdArgs = append(cmdArgs, "-c", c)
		}
	}
	cmdArgs = append(cmdArgs, pushCmd.Args...)
	envs := []string{}
	if len(pushCmd.Env) > 0 {
		envs = append(envs, pushCmd.Env...)
		if gitURL.IsSSH() {
			var sshCmdArgs []string
			sshCmd := helper.NewSSHCmd()
			sshCmdArgs, envs = sshCmd.Command("", 0, envs)
			shellCmd := helper.NewShellCmdFromArgs(sshCmdArgs...)
			envs = append(envs, "GIT_SSH_COMMAND="+shellCmd.QuoteCommand())
		}
	}

	if config.IsDryRun() || o.MockGitPush {
		log.Notef("%swill execute command: %s",
			v.Project.Prompt(),
			strings.Join(cmdArgs, " "))
		for _, env := range envs {
			log.Notef("%swith extra environment: %s", v.Project.Prompt(), env)
		}
	} else {
		log.Debugf("%sreview by command: %s",
			v.Project.Prompt(),
			strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = p.WorkDir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if len(envs) > 0 {
			cmd.Env = []string{}
			cmd.Env = append(cmd.Env, os.Environ()...)
			cmd.Env = append(cmd.Env, envs...)
			for _, env := range envs {
				log.Debugf("%swith extra environment: %s", v.Project.Prompt(), env)
			}
		}
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("upload failed: %s", err)
		}
	}

	branchName := v.Branch.Name
	if strings.HasPrefix(branchName, config.RefsHeads) {
		branchName = strings.TrimPrefix(branchName, config.RefsHeads)
	}

	var (
		publishedRef string
		msg          string
	)
	if v.CodeReview.Empty() {
		publishedRef = config.RefsPub + branchName
		msg = fmt.Sprintf("review from %s to %s on %s",
			branchName,
			o.DestBranch,
			v.Remote.Review)
	} else {
		publishedRef = v.CodeReview.Ref
		msg = fmt.Sprintf("update code review #%s of %s",
			v.CodeReview.ID,
			v.Remote.Review)
	}
	log.Debugf("%sUpdate reference '%s': %s",
		v.Project.Prompt(),
		publishedRef,
		msg)
	err = p.UpdateRef(publishedRef,
		config.RefsHeads+branchName,
		msg)

	if err != nil {
		return fmt.Errorf("fail to create or update reference '%s': %s",
			publishedRef,
			err)
	}
	return nil
}

// GetUploadableBranch returns branch which has commits ready for upload.
func (v *Project) GetUploadableBranch(branch string, remote *Remote, remoteBranch string) *ReviewableBranch {
	if remote == nil {
		log.Warnf("BUG: remote is nil for branch '%s' of project '%s'",
			branch,
			v.Name,
		)
		return nil
	}

	if branch == "" {
		branch = v.GetHead()
		if branch == "" {
			return nil
		}
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)

	if remoteBranch == "" {
		remoteBranch = v.Config().Get("branch." + branch + ".merge")
	}

	branchID, err := v.ResolveRevision(branch)
	if err != nil {
		return nil
	}
	track := v.RemoteMatchingBranch(remote.Name, remoteBranch)
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
			Remote: remote.Name,
			Branch: remoteBranch,
			Track: Reference{
				Name: track,
				Hash: trackID,
			},
		},
		Remote: remote,
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

// GetUploadableBranchForChange returns branch which has commits ready for upload.
func (v *Project) GetUploadableBranchForChange(branch string, remote *Remote, codeReview *common.CodeReview) *ReviewableBranch {
	if branch == "" {
		branch = v.GetHead()
		if branch == "" {
			return nil
		}
	}
	branch = strings.TrimPrefix(branch, config.RefsHeads)
	branchID, err := v.ResolveRevision(branch)
	if err != nil {
		return nil
	}

	rb := ReviewableBranch{
		Project: v,
		Branch: Branch{
			Name: branch,
			Hash: branchID},
		CodeReview: *codeReview,
		Remote:     remote,
	}

	if len(rb.Commits()) == 0 {
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
		remote := v.GetBranchRemote(branch, false)
		rb := v.GetUploadableBranch(branch, remote, "")
		if rb == nil {
			return nil
		}
		avail = append(avail, *rb)
		return avail
	}

	for _, head := range v.Heads() {
		remote := v.GetBranchRemote(branch, false)
		rb := v.GetUploadableBranch(head.Name, remote, "")
		if rb == nil {
			continue
		}
		avail = append(avail, *rb)
	}

	return avail
}
