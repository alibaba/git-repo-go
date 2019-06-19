package project

import (
	"fmt"
	"strings"

	log "github.com/jiangxin/multi-log"
)

// PatchSet contains code review reference and commits
type PatchSet struct {
	Reference string
	Commits   []string
	Commit    string
}

// DownloadPatchSet will fetch code review and return downloaded PatchSet
func (v Project) DownloadPatchSet(reviewID, patchID int) (*PatchSet, error) {
	reviewRef := ""
	if v.Remote != nil {
		reviewRef = v.Remote.GetCodeReviewRef(reviewID, patchID)
	}
	if reviewRef == "" {
		return nil, fmt.Errorf("cannot find review reference for %s", v.Name)
	}

	remote := v.Remote.GetRemote()
	if remote == nil {
		return nil, fmt.Errorf("unknown remote defined for %s", v.Name)
	}

	cmdArgs := []string{
		GIT,
		"fetch",
		remote.Name,
		"+" + reviewRef + ":" + reviewRef,
		"--",
	}
	log.Debugf("will execute: %s", strings.Join(cmdArgs, " "))
	err := executeCommandIn(v.WorkDir, cmdArgs)
	if err != nil {
		return nil, err
	}

	commits, err := v.Revlist(reviewRef, "--not", "HEAD")
	if err != nil {
		return nil, err
	}

	dl := PatchSet{
		Reference: reviewRef,
		Commits:   commits,
	}

	if len(dl.Commits) > 0 {
		dl.Commit = dl.Commits[0]
	} else {
		commit, err := v.ResolveRevision(reviewRef)
		if err != nil {
			return nil, err
		}
		dl.Commit = commit
	}

	return &dl, nil
}

// CherryPick will run cherry-pick on commits
func (v Project) CherryPick(commits ...string) error {
	for i := len(commits) - 1; i >= 0; i-- {
		c := commits[i]
		cmdArgs := []string{
			GIT,
			"cherry-pick",
			c,
			"--",
		}
		log.Debugf("will execute: %s", strings.Join(cmdArgs, " "))
		err := executeCommandIn(v.WorkDir, cmdArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Revert will run revert on commit
func (v Project) Revert(commit string) error {
	cmdArgs := []string{
		GIT,
		"revert",
		commit,
		"--",
	}
	log.Debugf("will execute: %s", strings.Join(cmdArgs, " "))
	return executeCommandIn(v.WorkDir, cmdArgs)
}
