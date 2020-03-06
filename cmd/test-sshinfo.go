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

	"github.com/alibaba/git-repo-go/helper"
	"github.com/spf13/cobra"
)

type testSSHInfoCommand struct {
	cmd *cobra.Command
}

func (v *testSSHInfoCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "sshinfo <connection>",
		Short: "test sshinfo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	return v.cmd
}

func (v testSSHInfoCommand) Execute(args []string) error {
	if len(args) != 1 {
		if len(args) == 0 {
			return fmt.Errorf("connection is not given in args")
		}
		if len(args) > 1 {
			return fmt.Errorf("only one connection (HTTP/SSH) should be given")
		}
	}

	query := helper.NewSSHInfoQuery("")
	sshInfo, err := query.GetSSHInfo(args[0], false)
	if err != nil {
		return err
	}
	fmt.Printf("ssh_info: %#v\n", sshInfo)

	return nil
}

var testSSHInfoCmd = testSSHInfoCommand{}

func init() {
	testCmd.Command().AddCommand(testSSHInfoCmd.Command())
}
