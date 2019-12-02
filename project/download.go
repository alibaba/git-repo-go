package project

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/jiangxin/multi-log"
)

// PatchSet contains code review reference and commits.
type PatchSet struct {
	Reference string
	Commits   []string
	Commit    string
}

// DownloadPatchSet fetches code review and return the downloaded PatchSet.
func (v Project) DownloadPatchSet(reviewID, patchID int) (*PatchSet, error) {
	if !v.Remote.Initialized() {
		log.Fatalf("%snot remote tracking defined, and do not know where to download",
			v.Prompt())
	}
	reviewRef, err := v.Remote.GetDownloadRef(strconv.Itoa(reviewID), strconv.Itoa(patchID))
	if err != nil {
		return nil, err
	}
	if reviewRef == "" {
		return nil, fmt.Errorf("cannot find review reference for %s", v.Name)
	}

	cmdArgs := []string{
		GIT,
		"fetch",
		v.Remote.Name,
		"+" + reviewRef + ":" + reviewRef,
		"--",
	}
	log.Debugf("%swill execute: %s", v.Prompt(), strings.Join(cmdArgs, " "))
	err = executeCommandIn(v.WorkDir, cmdArgs)
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

// CherryPick runs cherry-pick on commits.
func (v Project) CherryPick(commits ...string) error {
	for i := len(commits) - 1; i >= 0; i-- {
		c := commits[i]
		cmdArgs := []string{
			GIT,
			"cherry-pick",
			c,
			"--",
		}
		log.Debugf("%swill execute: %s", v.Prompt(), strings.Join(cmdArgs, " "))
		err := executeCommandIn(v.WorkDir, cmdArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

// Revert runs revert on commit.
func (v Project) Revert(commit string) error {
	cmdArgs := []string{
		GIT,
		"revert",
		commit,
		"--",
	}
	log.Debugf("%swill execute: %s", v.Prompt(), strings.Join(cmdArgs, " "))
	return executeCommandIn(v.WorkDir, cmdArgs)
}
