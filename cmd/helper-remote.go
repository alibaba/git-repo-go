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
	"io/ioutil"
	"os"
	"strings"

	"code.alibaba-inc.com/force/git-repo/helper"
	"github.com/spf13/cobra"
)

type helperRemoteCommand struct {
	cmd *cobra.Command
	O   struct {
		Upload   bool
		Download bool
		Type     string
	}
}

func (v *helperRemoteCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "remote",
		Short: "execute remote helper",

		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().StringVar(&v.O.Type,
		"type",
		"",
		"type of remote")
	v.cmd.Flags().BoolVar(&v.O.Upload,
		"upload",
		false,
		"output JSON for git upload commands")
	v.cmd.Flags().BoolVar(&v.O.Download,
		"download",
		false,
		"output JSON for download git reference")

	return v.cmd
}

func (v *helperRemoteCommand) Execute(arts []string) error {
	var (
		buf          []byte
		err          error
		ref          string
		remoteHelper helper.RemoteHelper
	)

	if v.O.Type == "" {
		return fmt.Errorf("must provide type of remote")
	}
	remoteHelper = helper.NewRemoteHelper(v.O.Type)

	if v.O.Download && v.O.Upload {
		return fmt.Errorf("cannot use --download and --upload together")
	}

	if v.O.Download {
		buf, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		slices := strings.SplitN(strings.TrimSpace(string(buf)), " ", 2)
		if len(slices) == 2 {
			ref, err = remoteHelper.GetDownloadRef(slices[0], slices[1])
		} else {
			ref, err = remoteHelper.GetDownloadRef(slices[0], "")
		}
		if err != nil {
			return err
		}
		fmt.Println(ref)
		return nil
	}

	buf, err = remoteHelper.GetGitPushCommandPipe(os.Stdin)
	if err != nil {
		return err
	}
	fmt.Println(string(buf))
	return nil
}

var helperRemoteCmd = helperRemoteCommand{}

func init() {
	helperCmd.Command().AddCommand(helperRemoteCmd.Command())
}
