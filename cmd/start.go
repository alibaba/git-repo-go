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
	"github.com/alibaba/git-repo-go/project"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type startCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   struct {
		All bool
	}
}

func (v *startCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "start",
		Short: "Start a new branch for development",
		Long:  `Begin a new branch of development, starting from the revision specified in the manifest.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().BoolVar(&v.O.All,
		"all",
		false,
		"begin branch in all projects")

	return v.cmd
}

func (v startCommand) Execute(args []string) error {
	var (
		failed    = []string{}
		execError error
	)

	rws := v.RepoWorkSpace()

	if len(args) == 0 {
		return newUserError("no args")
	}

	branch := args[0]

	names := []string{}
	if !v.O.All {
		if len(args) > 1 {
			names = append(names, args[1:]...)
		} else {
			// current project
			names = append(names, ".")
		}
	}

	allProjects, err := rws.GetProjects(nil, names...)
	if err != nil {
		return err
	}

	for _, p := range allProjects {
		merge := ""
		if project.IsImmutable(p.Revision) {
			if p.DestBranch != "" {
				merge = p.DestBranch
			} else {
				if rws.Manifest != nil &&
					rws.Manifest.Default != nil {
					merge = rws.Manifest.Default.Revision
				}
			}
		}
		err := p.StartBranch(branch, merge, false)
		if err != nil {
			failed = append(failed, p.Path)
			execError = err
		}
	}

	if execError != nil {
		for _, p := range failed {
			log.Errorf("cannot start branch '%s' for '%s'", branch, p)
		}
		return execError
	}
	return nil
}

var startCmd = startCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: false,
		SingleOK: false,
	},
}

func init() {
	rootCmd.AddCommand(startCmd.Command())
}
