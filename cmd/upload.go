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
	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type uploadCommand struct {
	cmd *cobra.Command
	ws  *workspace.WorkSpace

	O struct {
		AutoTopic     bool
		Reviewers     []string
		Cc            []string
		Branch        string
		CurrentBranch bool
		Draft         bool
		NoEmails      bool
		Private       bool
		WIP           bool
		PushOptions   []string
		DestBranch    string
		BypassHooks   bool
		AllowAllHooks bool
	}
}

func (v *uploadCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload changes for code review",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runE(args)
		},
	}

	aliasNormalizeFunc := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "re":
			name = "reviewers"
		case "reviewer":
			name = "reviewers"
		case "current-branch":
			name = "cbr"
		case "destination":
			name = "dest"
		}
		return pflag.NormalizedName(name)
	}

	v.cmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)
	v.cmd.Flags().BoolVarP(&v.O.AutoTopic,
		"auto-topic",
		"t",
		false,
		"Send local branch name for Code Review")
	v.cmd.Flags().StringArrayVar(&v.O.Reviewers,
		"reviewers",
		nil,
		"Request reviews from these people")
	v.cmd.Flags().StringArrayVar(&v.O.Cc,
		"cc",
		nil,
		"Also send email to these email addresses")
	v.cmd.Flags().StringVar(&v.O.Branch,
		"br",
		"",
		"Branch to upload")
	v.cmd.Flags().BoolVar(&v.O.CurrentBranch,
		"cbr",
		false,
		"Upload current git branch")
	v.cmd.Flags().BoolVarP(&v.O.Draft,
		"draft",
		"d",
		false,
		"If specified, upload as a draft")
	v.cmd.Flags().BoolVar(&v.O.NoEmails,
		"no-emails",
		false,
		"If specified, do not send emails on upload")
	v.cmd.Flags().BoolVarP(&v.O.Private,
		"private",
		"p",
		false,
		"If specified, upload as a private change")
	v.cmd.Flags().BoolVarP(&v.O.WIP,
		"wip",
		"w",
		false,
		"If specified, upload as a work-in-progress change")
	v.cmd.Flags().StringArrayVarP(&v.O.PushOptions,
		"push-options",
		"o",
		nil,
		"Additional push options to transmit")
	v.cmd.Flags().StringVarP(&v.O.DestBranch,
		"dest",
		"D",
		"",
		"Submit for review on this target branch")
	v.cmd.Flags().Bool("no-cert-checks",
		false,
		"Disable verifying ssl certs (unsafe)")
	v.cmd.Flags().BoolVar(&v.O.BypassHooks,
		"no-verify",
		false,
		"Do not run the upload hook")
	v.cmd.Flags().BoolVar(&v.O.AllowAllHooks,
		"verify",
		false,
		"Run the upload hook without prompting")

	v.cmd.Flags().MarkHidden("auto-topic")

	viper.BindPFlag("no-cert-checks", v.cmd.Flags().Lookup("no-cert-checks"))

	return v.cmd

}

func (v *uploadCommand) WorkSpace() *workspace.WorkSpace {
	if v.ws == nil {
		v.reloadWorkSpace()
	}
	return v.ws
}

func (v *uploadCommand) reloadWorkSpace() {
	var err error
	v.ws, err = workspace.NewWorkSpace("")
	if err != nil {
		log.Fatal(err)
	}
}

func (v uploadCommand) runE(args []string) error {
	var (
		failed    = []string{}
		execError error
	)

	ws := v.WorkSpace()
	err := ws.LoadRemotes()
	if err != nil {
		return err
	}

	allProjects, err := ws.GetProjects(nil, args...)
	if err != nil {
		return err
	}

	for _, p := range allProjects {
		if v.O.CurrentBranch {
			cbr := p.GetHead()
			uploadBranch := p.GetUploadableBranch(cbr)
			log.Debugf("uploadBranch is : %s", uploadBranch)
		}
	}

	_, _, _ = failed, execError, ws

	log.Notef("reviewers: %#v", v.O.Reviewers)
	log.Notef("cc: %#v", v.O.Cc)

	return nil
}

var uploadCmd = uploadCommand{}

func init() {
	rootCmd.AddCommand(uploadCmd.Command())
}
