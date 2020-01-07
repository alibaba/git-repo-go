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
	"errors"
	"fmt"
	"strings"

	"github.com/aliyun/git-repo-go/color"
	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/path"
	"github.com/aliyun/git-repo-go/project"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type statusCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   struct {
		Jobs    int
		Orphans bool
	}
}

func (v *statusCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "status",
		Short: "Show the working tree status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().BoolVarP(&v.O.Orphans,
		"orphans",
		"o",
		false,
		"include objects in working directory outside of repo projects")
	v.cmd.Flags().IntVarP(&v.O.Jobs,
		"jobs",
		"j",
		2,
		"number of projects to check simultaneously")

	return v.cmd
}

func (v statusCommand) Execute(args []string) error {
	var (
		projects []*project.Project
		err      error
	)

	ws := v.RepoWorkSpace()

	if v.O.Jobs < 1 {
		v.O.Jobs = 1
	}

	projects, err = ws.GetProjects(nil, args...)
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		log.Infof("no projects")
		return nil
	}

	return v.RunCommand(projects)
}

func (v statusCommand) RunCommand(projects []*project.Project) error {
	var (
		jobs       = v.O.Jobs
		jobTasks   = make(chan int, jobs)
		jobResults = make(chan *project.CmdExecResult, jobs)
	)

	worker := func(i int) {
		log.Debugf("start command worker #%d", i)
		for idx := range jobTasks {
			jobResults <- v.executeCommand(projects[idx])
		}
	}

	for i := 0; i < jobs; i++ {
		go worker(i)
	}

	go func() {
		for i := 0; i < len(projects); i++ {
			jobTasks <- i
		}
		close(jobTasks)
	}()

	isClean := true
	count := len(projects)
	for i := 0; i < count; i++ {
		result := <-jobResults
		if result == nil {
			continue
		}
		if !result.Empty() {
			isClean = false
		}
		v.showResult(result, i, count)
	}

	if isClean {
		log.Note("nothing to commit (working directory clean)")
	}

	// TODO: handle --orphans option

	return nil
}

func (v statusCommand) showResult(result *project.CmdExecResult, i, count int) {
	stdout := result.Stdout()
	stderr := result.Stderr()
	if stdout == "" && stderr == "" {
		return
	}

	branchName := ""
	projectHeader := ""
	if result.Project != nil {
		branchName = result.Project.GetHead()
		if result.Project.Settings.Mirror {
			projectHeader = result.Project.Name
		} else {
			projectHeader = result.Project.Path
		}
	}
	fmt.Printf("%sproject %-40s%s",
		color.Color("normal", "", "bold"),
		projectHeader+"/ ",
		color.Reset())

	if branchName == "" {
		fmt.Printf("%s(*** NO BRANCH ***)%s",
			color.Color("red", "", "normal"),
			color.Reset(),
		)
	} else {
		branchName = strings.TrimPrefix(branchName, config.RefsHeads)
		fmt.Printf("%sbranch %s%s",
			color.Color("normal", "", "bold"),
			branchName,
			color.Reset(),
		)
	}

	fmt.Print("\n")

	if stdout != "" {
		if stdout[len(stdout)-1] != '\n' {
			fmt.Println(stdout)
		} else {
			fmt.Print(stdout)
		}
	}

	if stderr != "" {
		if stderr[len(stderr)-1] != '\n' {
			fmt.Println(stderr)
		} else {
			fmt.Print(stderr)
		}
	}

	if i < count-1 {
		fmt.Print("\n")
	}
}

func (v statusCommand) executeCommand(p *project.Project) *project.CmdExecResult {
	if !path.Exist(p.WorkDir) {
		result := project.CmdExecResult{}
		result.Project = p
		result.Error = errors.New(`missing (run "git repo sync")`)
		return &result
	}

	return p.Status()
}

var statusCmd = statusCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: false,
		SingleOK: false,
	},
}

func init() {
	rootCmd.AddCommand(statusCmd.Command())
}
