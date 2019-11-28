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
	"github.com/spf13/cobra"
)

type helperCommand struct {
	cmd *cobra.Command
}

func (v *helperCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}
	v.cmd = &cobra.Command{
		Use:    "helper <subcommand>",
		Short:  "verious helper program for git-repo",
		Hidden: true,
	}
	return v.cmd
}

var helperCmd = helperCommand{}

func init() {
	rootCmd.AddCommand(helperCmd.Command())
}
