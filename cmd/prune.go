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
	"sort"
	"strings"

	"code.alibaba-inc.com/force/git-repo/color"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/project"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type pruneCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   struct {
		All    bool
		Branch string
		Force  bool
	}
}

func (v *pruneCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "prune [<project>...]",
		Short: "Prune (delete) already merged topic branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	return v.cmd
}

type projectBranch struct {
	project.Branch

	Project   *project.Project
	IsCurrent bool
}

type pbByBranch []projectBranch

func (v pbByBranch) Len() int      { return len(v) }
func (v pbByBranch) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v pbByBranch) Less(i, j int) bool {
	if v[i].Name == v[j].Name {
		return v[i].Project.Path < v[j].Project.Path
	}
	return v[i].Name < v[j].Name
}

type pbByPath []projectBranch

func (v pbByPath) Len() int      { return len(v) }
func (v pbByPath) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v pbByPath) Less(i, j int) bool {
	if v[i].Project.Path == v[j].Project.Path {
		return v[i].Name < v[j].Name
	}
	return v[i].Project.Path < v[j].Project.Path
}

func (v pruneCommand) Execute(args []string) error {
	var (
		err     error
		success = []projectBranch{}
		failure = []projectBranch{}
	)

	ws := v.WorkSpace()
	err = ws.LoadRemotes(false)
	if err != nil {
		return err
	}

	projects, err := ws.GetProjects(nil, args...)
	if err != nil {
		return err
	}

	if v.O.Branch != "" && !strings.HasPrefix(v.O.Branch, config.RefsHeads) {
		v.O.Branch = config.RefsHeads + v.O.Branch
	}

	maxProjectWidth := 0
	for _, p := range projects {
		w := len(p.Path)
		if w > maxProjectWidth {
			maxProjectWidth = w
		}
		allHeads := p.Heads()
		branches := []project.Branch{}
		cb := p.HeadBranch()
		if v.O.Branch != "" && v.O.Branch != cb.Name {
			oid, err := p.ResolveRevision(v.O.Branch)
			if err != nil {
				log.Debugf("fail to resolve %s in %s", v.O.Branch, p.Path)
			} else {
				branches = append(branches,
					project.Branch{
						Name: v.O.Branch,
						Hash: oid,
					})
			}
		} else if v.O.All {
			for _, b := range allHeads {
				if b.Name != cb.Name {
					branches = append(branches, b)
				}
			}
		}

		// Check current branch for deletion
		if cb.Name != "" && (v.O.All || cb.Name == v.O.Branch) {
			if v.O.Force {
				err := p.DetachHead()
				if err != nil {
					log.Errorf("cannot detach current branch for abandon: %s", err)
				} else {
					branches = append(branches, cb)
				}
			} else {
				// Check if there are no commits ahead for current branch
				tb := p.LocalTrackBranch(cb.Name)
				if tb == "" {
					tb, err = p.ResolveRemoteTracking(cb.Name)
				} else {
					tb, err = p.ResolveRevision(tb)
				}
				if err != nil {
					log.Errorf("cannot find tracking branch %s of %s: %s",
						cb.Name,
						p.Path,
						err)
				} else {
					list, err := p.Revlist(cb.Name, "--not", tb)
					if err != nil {
						log.Errorf("revlist failed for HEAD: %s", err)
					} else if len(list) == 0 && p.IsClean() {
						err := p.DetachHead()
						if err != nil {
							log.Errorf("cannot detach current branch for prune: %s", err)
						} else {
							branches = append(branches, cb)
						}
					} else {
						failure = append(failure, projectBranch{
							Branch:    cb,
							Project:   p,
							IsCurrent: true,
						})
					}
				}
			}
		}

		needToClean := false
		if len(branches) == 0 {
			log.Debugf("no branch to prune for project %s", p.Name)
		} else {
			// run bfranch -d ...
			cmdArgs := []string{
				project.GIT,
				"branch",
			}
			if v.O.Force {
				cmdArgs = append(cmdArgs, "-D")
			} else {
				cmdArgs = append(cmdArgs, "-d")
			}
			for _, b := range branches {
				branchName := strings.TrimPrefix(b.Name, config.RefsHeads)
				cmdArgs = append(cmdArgs, branchName)
			}

			result := p.ExecuteCommand(cmdArgs...)
			if result.Error == nil {
				needToClean = true
				for _, b := range branches {
					success = append(success, projectBranch{
						Branch:    b,
						Project:   p,
						IsCurrent: b.Name == cb.Name,
					})
				}
			} else {
				left := make(map[string]bool)
				for _, b := range p.Heads() {
					left[b.Name] = true
				}
				for _, b := range branches {
					if _, ok := left[b.Name]; ok {
						failure = append(failure, projectBranch{
							Branch:    b,
							Project:   p,
							IsCurrent: b.Name == cb.Name,
						})
					} else {
						needToClean = true
						success = append(success, projectBranch{
							Branch:  b,
							Project: p,
						})
					}
				}
			}
		}

		if needToClean {
			log.Debugf("clean published cache for %s", p.Path)
			p.CleanPublishedCache()
		}
	}

	// Show deleted branch
	if len(success) > 0 {
		if v.O.Force {
			color.Hilightln("Abandoned branches")
		} else {
			color.Hilightln("Pruned branches (already merged)")
		}
		color.Dimln(strings.Repeat("-", 78))
		sort.Sort(pbByBranch(success))
		maxBranchWidth := 25
		for _, b := range success {
			w := len(b.Name) - 11
			if w > maxBranchWidth {
				maxBranchWidth = w
			}
		}
		branchName := ""
		for _, b := range success {
			if config.RefsHeads+branchName != b.Name {
				if branchName != "" {
					fmt.Printf("\n")
				}
				branchName = strings.TrimPrefix(b.Name, config.RefsHeads)
				fmt.Printf("%-*s | ", maxBranchWidth, branchName)
			} else {
				fmt.Printf("%-*s | ", maxBranchWidth, "")
			}
			fmt.Printf("%-*s (was %s)\n", maxProjectWidth, b.Project.Path, b.Hash[0:7])
		}
		fmt.Println("")
	}

	// Show what's remaining, which has unmerged commits
	if len(failure) > 0 {
		color.Hilightln("Pending branches (which have unmerged commits, leave it as is)")
		color.Dimln(strings.Repeat("-", 78))
		sort.Sort(pbByPath(failure))
		projectPath := ""
		for _, b := range failure {
			p := b.Project
			if projectPath != p.Path {
				if projectPath != "" {
					fmt.Println("")
				}
				projectPath = p.Path
				fmt.Printf("Project %s/\n", projectPath)
			}

			if b.IsCurrent {
				fmt.Print("* ")
			} else {
				fmt.Print("  ")
			}
			branchName := strings.TrimPrefix(b.Name, config.RefsHeads)
			remote := p.GetBranchRemote(branchName, false)
			if remote == nil {
				fmt.Printf("%s (no remote)\n", branchName)
			}
			rb := p.GetUploadableBranch(b.Name, remote, "")
			if rb != nil {
				commits := rb.Commits()
				if len(commits) > 1 {
					fmt.Printf("%s (%2d commits, %s)\n",
						branchName,
						len(commits),
						p.LastModified(b.Hash),
					)
				} else {
					fmt.Printf("%s (%2d commit, %s)\n",
						branchName,
						len(commits),
						p.LastModified(b.Hash),
					)
				}
			} else {
				fmt.Printf("%s\n", branchName)
			}
		}
	}

	return nil
}

var pruneCmd = pruneCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: false,
		SingleOK: true,
	},
}

func init() {
	pruneCmd.O.All = true
	pruneCmd.O.Force = false
	rootCmd.AddCommand(pruneCmd.Command())
}
