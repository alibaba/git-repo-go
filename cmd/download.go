// Copyright © 2019 Alibaba Co. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"regexp"
	"strconv"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/project"
	"code.alibaba-inc.com/force/git-repo/workspace"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type downloadCommand struct {
	cmd *cobra.Command
	ws  workspace.WorkSpace

	O struct {
		CherryPick bool
		Revert     bool
		FFOnly     bool
		NoCache    bool
	}
}

// projectChange wraps download project and review ID
type projectChange struct {
	Project  *project.Project
	ReviewID int
	PatchID  int
}

var (
	reChange = regexp.MustCompile(`^([1-9][0-9]*)(?:[/\.-]([1-9][0-9]*))?$`)
)

func (v *downloadCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "download",
		Short: "Download and checkout a code review",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runE(args)
		},
	}
	v.cmd.Flags().BoolVarP(&v.O.CherryPick,
		"cherry-pick",
		"c",
		false,
		"cherry-pick instead of checkout")
	v.cmd.Flags().BoolVarP(&v.O.Revert,
		"revert",
		"r",
		false,
		"revert instead of checkout")
	v.cmd.Flags().BoolVarP(&v.O.FFOnly,
		"ff-only",
		"f",
		false,
		"force fast-forward merge")
	v.cmd.Flags().BoolVar(&v.O.NoCache,
		"no-cache",
		false,
		"Ignore ssh-info cache, and recheck ssh-info API")

	return v.cmd
}

func (v *downloadCommand) parseChanges(args ...string) ([]projectChange, error) {
	var (
		changes []projectChange
		p       *project.Project
	)

	for _, arg := range args {
		matches := reChange.FindStringSubmatch(arg)
		if matches == nil || p == nil {
			projectName := arg
			if matches != nil {
				projectName = "."
			}
			projects, err := v.ws.GetProjects(nil, projectName)
			if err != nil {
				return nil, err
			}
			if len(projects) == 0 {
				return nil, fmt.Errorf("cannot find project matched for '%s'", projectName)
			}
			p = projects[0]
			if matches == nil {
				continue
			}
		}

		pr := projectChange{Project: p}
		pr.ReviewID, _ = strconv.Atoi(matches[1])
		if len(matches) >= 3 {
			pr.PatchID, _ = strconv.Atoi(matches[2])
		}
		changes = append(changes, pr)
	}
	return changes, nil
}

func (v *downloadCommand) WorkSpace() workspace.WorkSpace {
	if v.ws == nil {
		v.reloadWorkSpace()
	}
	return v.ws
}

func (v *downloadCommand) reloadWorkSpace() {
	var err error
	v.ws, err = workspace.NewWorkSpace("")
	if err != nil {
		log.Fatal(err)
	}
}

func (v *downloadCommand) runE(args []string) error {
	ws := v.WorkSpace()
	err := ws.LoadRemotes(v.O.NoCache)
	if err != nil {
		return err
	}

	n := 0
	if v.O.CherryPick {
		n++
	}
	if v.O.Revert {
		n++
	}
	if v.O.FFOnly {
		n++
	}
	if n > 1 {
		return fmt.Errorf("cannot use more than one of `-c`, `-r`, or `-f` options")
	}

	if len(args) == 0 {
		return newUserError("no args")
	}

	changes, err := v.parseChanges(args...)
	if err != nil {
		return err
	}

	for _, c := range changes {
		dl, err := c.Project.DownloadPatchSet(c.ReviewID, c.PatchID)
		if err != nil {
			return err
		}

		changeID := ""
		if c.PatchID == 0 {
			changeID = fmt.Sprintf("%d", c.ReviewID)
		} else {
			changeID = fmt.Sprintf("%d/%d", c.ReviewID, c.PatchID)
		}

		if len(dl.Commits) == 0 && !v.O.Revert {
			log.Notef("[%s] change %s has already been merged",
				c.Project.Name, changeID)
			continue
		}

		if len(dl.Commits) > 1 {
			log.Notef("[%s] %s depends on %d unmerged changes:",
				c.Project.Name,
				changeID,
				len(dl.Commits))
			for _, commit := range dl.Commits {
				log.Notef("  %s", commit)
			}
		}

		if v.O.CherryPick {
			answer := true
			if len(dl.Commits) > unusualCommitThreshold {
				input := userInput(
					fmt.Sprintf("Too many commits(%d) to cherry pick, are you sure (y/N)? ", len(dl.Commits)),
					"N",
				)
				if !answerIsTrue(input) {
					answer = false
				}
			}

			if answer {
				err = c.Project.CherryPick(dl.Commits...)
			} else {
				err = fmt.Errorf("cherry-pick aborted by user")
			}
		} else if v.O.Revert {
			if c.Project.Remote != nil && c.Project.Remote.GetType() == config.RemoteTypeGerrit {
				err = c.Project.Revert(dl.Commit)
			} else {
				err = fmt.Errorf("--revert only works for gerrit server")
			}
		} else if v.O.FFOnly {
			err = c.Project.FastForward("--ff-only", dl.Commit)
		} else {
			err = c.Project.CheckoutRevision(dl.Commit)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

var downloadCmd = downloadCommand{}

func init() {
	rootCmd.AddCommand(downloadCmd.Command())
}
