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
	"sort"

	"github.com/alibaba/git-repo-go/project"
	"github.com/alibaba/git-repo-go/workspace"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type listCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   struct {
		Regex    []string
		Groups   string
		FullPath bool
		NameOnly bool
		PathOnly bool
	}
}

func (v *listCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "list",
		Short: "List projects and their associated directories",
		Long: `List all projects; pass '.' to list the project for the cwd.
		This is similar to running: git-repo forall -c 'echo "$REPO_PATH :
		$REPO_PROJECT"'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().StringArrayVarP(&v.O.Regex,
		"regex",
		"r",
		nil,
		"Filter the project list based on regex or wildcard matching of strings")
	v.cmd.Flags().StringVarP(&v.O.Groups,
		"groups",
		"g",
		"",
		"Filter the project list based on the groups the project is in")
	v.cmd.Flags().BoolVarP(&v.O.FullPath,
		"fullpath",
		"f",
		false,
		"Display the full work tree path instead of the relative path")
	v.cmd.Flags().BoolVarP(&v.O.NameOnly,
		"name-only",
		"n",
		false,
		"Display only the name of the repository")
	v.cmd.Flags().BoolVarP(&v.O.PathOnly,
		"path-only",
		"p",
		false,
		"Display only the path of the repository")

	return v.cmd
}

func (v listCommand) Execute(args []string) error {
	var (
		projects    []*project.Project
		allProjects []*project.Project
		patterns    []*regexp.Regexp
		err         error
	)

	ws := v.RepoWorkSpace()

	if v.O.NameOnly && v.O.PathOnly {
		log.Fatal("cannot combine -p and -n")
	}

	if v.O.NameOnly && v.O.FullPath {
		log.Fatal("cannot combine -f and -n")
	}

	allProjects, err = ws.GetProjects(&workspace.GetProjectsOptions{
		Groups: v.O.Groups,
	})

	if err != nil {
		return err
	}

	for _, r := range v.O.Regex {
		re, err := regexp.Compile(r)
		if err != nil {
			log.Warnf("cannot compile regex pattern %s: %s", r, err)
			continue
		}
		patterns = append(patterns, re)
	}

	if len(patterns) == 0 {
		projects = allProjects
	} else {
		for _, p := range allProjects {
			for _, re := range patterns {
				if re.MatchString(p.Name) || re.MatchString(p.Path) {
					projects = append(projects, p)
					break
				}
			}
		}
	}

	if len(projects) == 0 {
		log.Notef("no projects")
		return nil
	}

	outputs := make([]string, len(projects))
	for i, project := range projects {
		if v.O.NameOnly {
			outputs[i] = fmt.Sprintf("%s\n", project.Name)
		} else if v.O.PathOnly {
			outputs[i] = fmt.Sprintf("%s\n", v.getPath(project))
		} else {
			outputs[i] = fmt.Sprintf("%s : %s\n", v.getPath(project), project.Name)
		}
	}

	sort.Strings(outputs)

	for _, output := range outputs {
		fmt.Print(output)
	}
	return nil
}

func (v listCommand) getPath(project *project.Project) string {
	if v.O.FullPath {
		return project.WorkDir
	}
	return project.Path
}

var listCmd = listCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: true,
		SingleOK: false,
	},
}

func init() {
	rootCmd.AddCommand(listCmd.Command())
}
