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

	"code.alibaba-inc.com/force/git-repo/workspace"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type manifestCommand struct {
	cmd *cobra.Command
	ws  *workspace.RepoWorkSpace

	O struct {
		PegRev           bool
		PegRevNoUpstream bool
		OutputFile       string
	}
}

func (v *manifestCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "manifest",
		Short: "Manifest inspection utility",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	v.cmd.Flags().BoolVarP(&v.O.PegRev,
		"revision-as-HEAD",
		"r",
		false,
		"Save revisions as current HEAD")
	v.cmd.Flags().BoolVar(&v.O.PegRevNoUpstream,
		"suppress-upstream-revision",
		false,
		"If in -r mode, do not write the upstream field.  "+
			"Only of use if the branch names for a sha1 "+
			"manifest are sensitive.")
	v.cmd.Flags().StringVarP(&v.O.OutputFile,
		"output-file",
		"o",
		"-",
		"File to save the manifest to")

	return v.cmd
}

func (v *manifestCommand) RepoWorkSpace() *workspace.RepoWorkSpace {
	var err error
	if v.ws == nil {
		v.ws, err = workspace.NewRepoWorkSpace("")
		if err != nil {
			log.Fatal(err)
		}
	}
	return v.ws
}

func (v manifestCommand) Execute(args []string) error {
	ws := v.RepoWorkSpace()
	_ = ws
	fmt.Println("manifest cmd test print.")
	return nil
}

var manifestCmd = manifestCommand{}

func init() {
	rootCmd.AddCommand(manifestCmd.Command())
}
