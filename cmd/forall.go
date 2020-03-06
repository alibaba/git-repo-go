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
	"os"
	"regexp"
	"strconv"

	"github.com/alibaba/git-repo-go/color"
	"github.com/alibaba/git-repo-go/path"
	"github.com/alibaba/git-repo-go/project"
	"github.com/alibaba/git-repo-go/workspace"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type forallCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   struct {
		AbortOnErrors bool
		InverseRegex  []string
		Regex         []string
		ProjectHeader bool
		Command       string
		Groups        string
		Jobs          int
	}
}

func (v *forallCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "forall",
		Short: "Run a shell command in each project",
		Long:  `Executes the same shell command in each project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().StringArrayVarP(&v.O.Regex,
		"regex",
		"r",
		nil,
		"Execute the command only on projects matching regex or wildcard expression")
	v.cmd.Flags().StringArrayVarP(&v.O.InverseRegex,
		"inverse-regex",
		"i",
		nil,
		"Execute the command only on projects not matching regex or wildcard expression")
	v.cmd.Flags().StringVarP(&v.O.Groups,
		"groups",
		"g",
		"",
		"Execute the command only on projects matching the specified groups")
	v.cmd.Flags().StringVarP(&v.O.Command,
		"command",
		"c",
		"",
		"Command (and arguments) to execute")
	v.cmd.Flags().BoolVarP(&v.O.AbortOnErrors,
		"abort-on-errors",
		"e",
		false,
		"Abort if a command exits unsuccessfully")
	v.cmd.Flags().BoolVarP(&v.O.ProjectHeader,
		"project-header",
		"p",
		false,
		"Show project headers before output")
	v.cmd.Flags().IntVarP(&v.O.Jobs,
		"jobs",
		"j",
		1,
		"number of commands to execute simultaneously")

	return v.cmd
}

func (v forallCommand) Execute(args []string) error {
	var (
		cmds        []string
		allProjects []*project.Project
		projects    []*project.Project
		err         error
		inverseMode bool
		patterns    []*regexp.Regexp
	)

	ws := v.RepoWorkSpace()

	if v.O.Command != "" {
		cmds = append(cmds, v.O.Command)
	}
	if len(args) > 0 {
		cmds = append(cmds, args...)
	}
	if len(cmds) == 0 {
		return fmt.Errorf("no command provided")
	}
	if v.O.Jobs < 1 {
		v.O.Jobs = 1
	}

	inverseMode = len(v.O.InverseRegex) > 0
	if inverseMode && len(v.O.Regex) > 0 {
		return fmt.Errorf("--regex and --inverse-regex cannot be used together")
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
	for _, r := range v.O.InverseRegex {
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
			matched := false
			for _, re := range patterns {
				if re.MatchString(p.Name) || re.MatchString(p.Path) {
					matched = true
					if !inverseMode {
						projects = append(projects, p)
					}
					break
				}
			}
			if inverseMode && !matched {
				projects = append(projects, p)
			}
		}
	}

	if len(projects) == 0 {
		log.Infof("no projects")
		return nil
	}

	return v.RunCommand(projects, cmds)
}

func (v forallCommand) RunCommand(projects []*project.Project, cmds []string) error {
	var (
		jobs       = v.O.Jobs
		jobTasks   = make(chan int, jobs)
		jobResults = make(chan *project.CmdExecResult, jobs)
	)

	os.Setenv("REPO_COUNT", strconv.Itoa(len(projects)))

	if !regexp.MustCompile(`^[a-z0-9A-Z_/\.-]+$`).MatchString(cmds[0]) {
		shellCmd := []string{
			"sh",
			"-c",
			cmds[0],
		}
		shellCmd = append(shellCmd, cmds...)
		cmds = shellCmd
	}

	worker := func(i int) {
		log.Debugf("start command worker #%d", i)
		for idx := range jobTasks {
			jobResults <- v.executeCommand(projects[idx], cmds)
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

	count := len(projects)
	for i := 0; i < count; i++ {
		result := <-jobResults
		if result == nil {
			continue
		}
		v.showResult(result, i, count)
	}
	return nil
}

func (v forallCommand) showResult(result *project.CmdExecResult, i, count int) {
	stdout := result.Stdout()
	stderr := result.Stderr()
	if stdout == "" && stderr == "" {
		return
	}

	if v.O.ProjectHeader {
		projectHeader := ""
		if result.Project != nil {
			if result.Project.Settings.Mirror {
				projectHeader = result.Project.Name
			} else {
				projectHeader = result.Project.Path
			}
		}
		fmt.Printf("%sproject %s/%s\n",
			color.Color("normal", "", "bold"),
			projectHeader,
			color.Reset())
	}

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

	if v.O.ProjectHeader && i < count-1 {
		fmt.Print("\n")
	}
}

func (v forallCommand) executeCommand(p *project.Project, cmds []string) *project.CmdExecResult {
	workdir := p.WorkDir
	if p.IsMirror() {
		workdir = p.GitDir
	}
	if !path.Exist(workdir) {
		log.Infof("skipping %s/", p.Path)
		return nil
	}

	os.Setenv("REPO_PROJECT", p.Name)
	os.Setenv("REPO_PATH", p.Path)
	os.Setenv("REPO_REMOTE", p.RemoteName)

	return p.ExecuteCommand(cmds...)
}

var forallCmd = forallCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: true,
		SingleOK: false,
	},
}

func init() {
	rootCmd.AddCommand(forallCmd.Command())
}
