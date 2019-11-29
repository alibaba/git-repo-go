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

	"code.alibaba-inc.com/force/git-repo/version"
	"github.com/spf13/cobra"
)

type testVersionCommand struct {
	cmd *cobra.Command

	O struct {
		TestGit bool
	}
}

func (v *testVersionCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "version",
		Short: "test version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	v.cmd.Flags().BoolVarP(&v.O.TestGit,
		"git",
		"g",
		false,
		"test git instead of git-repo")

	return v.cmd
}

func (v testVersionCommand) Execute(args []string) error {
	var (
		cmp           int
		err           error
		actualVersion string
	)

	if len(args) != 2 {
		return fmt.Errorf("Usage: git-repo test version <op> <version>")
	}

	if v.O.TestGit {
		actualVersion = version.GitVersion
	} else {
		actualVersion = version.Version
	}
	cmp = version.CompareVersion(actualVersion, args[1])

	switch args[0] {
	case "lt":
		if cmp >= 0 {
			err = fmt.Errorf("%s not little than %s", actualVersion, args[1])
		}
	case "le":
		if cmp > 0 {
			err = fmt.Errorf("%s not little or equal than %s", actualVersion, args[1])
		}
	case "gt":
		if cmp <= 0 {
			err = fmt.Errorf("%s not greater than %s", actualVersion, args[1])
		}
	case "ge":
		if cmp < 0 {
			err = fmt.Errorf("%s not greater or equal than %s", actualVersion, args[1])
		}
	case "eq":
		if cmp != 0 {
			err = fmt.Errorf("%s and %s are not equal", actualVersion, args[1])
		}
	}
	return err
}

var testVersionCmd = testVersionCommand{}

func init() {
	testCmd.Command().AddCommand(testVersionCmd.Command())
}
