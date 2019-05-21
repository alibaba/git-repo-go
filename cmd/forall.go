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
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"code.alibaba-inc.com/force/git-repo/color"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type forallCommand struct {
	cmd *cobra.Command
	ws  *workspace.RepoWorkSpace

	O struct {
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
			return v.runE(args)
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

func (v *forallCommand) RepoWorkSpace() *workspace.RepoWorkSpace {
	if v.ws == nil {
		v.reloadRepoWorkSpace()
	}
	return v.ws
}

func (v *forallCommand) reloadRepoWorkSpace() {
	var err error
	v.ws, err = workspace.NewRepoWorkSpace("")
	if err != nil {
		log.Fatal(err)
	}
}

func (v forallCommand) runE(args []string) error {
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
		jobResults = make(chan *forallResult, jobs)
	)

	os.Setenv("REPO_COUNT", strconv.Itoa(len(projects)))

	execShell := !regexp.MustCompile(`^[a-z0-9A-Z_/\.-]+$`).MatchString(cmds[0])
	worker := func(i int) {
		log.Debugf("start command worker #%d", i)
		for idx := range jobTasks {
			jobResults <- v.executeCommand(projects[idx], execShell, cmds)
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

	for i := 0; i < len(projects); i++ {
		result := <-jobResults
		if result == nil {
			continue
		}
		fmt.Print(result.Output)
		if v.O.ProjectHeader && i < len(projects)-1 {
			fmt.Print("\n")
		}
	}
	return nil
}

type forallResult struct {
	Project *project.Project
	Output  string
	Error   error
}

func (v forallCommand) executeCommand(p *project.Project, execShell bool, cmds []string) *forallResult {
	var (
		dir           string
		projectHeader string
		result        = forallResult{}
		isMirror      = v.ws.ManifestProject != nil && v.ws.ManifestProject.MirrorEnabled()
		cmd           *exec.Cmd
	)

	result.Project = p
	if isMirror {
		dir = p.WorkRepository.Path
	} else {
		dir = p.WorkDir
	}

	if !path.Exists(dir) {
		log.Infof("skipping %s/", p.Path)
		return nil
	}

	os.Setenv("REPO_PROJECT", p.Name)
	os.Setenv("REPO_PATH", p.Path)
	os.Setenv("REPO_REMOTE", p.RemoteName)

	if execShell {
		shellCmd := []string{
			"sh",
			"-c",
			cmds[0],
		}
		shellCmd = append(shellCmd, cmds...)
		cmd = exec.Command(shellCmd[0], shellCmd[1:]...)
		log.Debugf("execute: %s", strings.Join(shellCmd, " "))
	} else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
		log.Debugf("execute: %s", strings.Join(cmds, " "))
	}
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()

	if (len(out) > 0 || err != nil) && v.O.ProjectHeader {
		if isMirror {
			projectHeader = p.Name
		} else {
			projectHeader = p.Path
		}
		result.Output += fmt.Sprintf("%sproject %s/%s\n",
			color.Color("normal", "", "bold"),
			projectHeader,
			color.ColorReset())
	}

	if len(out) > 0 {
		result.Output += string(out)
		if result.Output[len(result.Output)-1] != '\n' {
			result.Output += "\n"
		}
	}
	if err != nil {
		result.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Output += string(exitErr.Stderr)
			if result.Output[len(result.Output)-1] != '\n' {
				result.Output += "\n"
			}
		}
	}

	return &result
}

var forallCmd = forallCommand{}

func init() {
	rootCmd.AddCommand(forallCmd.Command())
}
