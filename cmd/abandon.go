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

	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/spf13/cobra"
)

type abandonCommand struct {
	cmd *cobra.Command
	ws  *workspace.RepoWorkSpace

	O struct {
		All    bool
		Branch string
		Force  bool
	}
}

func (v *abandonCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "abandon [<project>...]",
		Short: "Permanently abandon a development branch with --force option",
		Long: `This subcommand permanently abandons a development branch by
deleting it from your local repository, if run with --force option.
It is equivalent to "git branch -D <branchname>".

Without --force option, only delete already merge branches, which
like command "git repo prune".`,
		Args: func(cmd *cobra.Command, args []string) error {
			if !v.O.All && v.O.Branch == "" {
				return errors.New("use --all or --branch to provide a branch to be abandoned")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().StringVarP(&v.O.Branch,
		"branch",
		"b",
		"",
		"delete specific branch")
	v.cmd.Flags().BoolVar(&v.O.All,
		"all",
		false,
		"delete all branches in all projects")
	v.cmd.Flags().BoolVar(&v.O.Force,
		"force",
		false,
		"delete branches even not published")

	return v.cmd
}

func (v abandonCommand) Execute(args []string) error {
	cmd := pruneCommand{}
	cmd.O = v.O

	return cmd.Execute(args)
}

var abandonCmd = abandonCommand{}

func init() {
	rootCmd.AddCommand(abandonCmd.Command())
}
