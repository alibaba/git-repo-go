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

	"github.com/spf13/cobra"
)

const (
	filterKindKeywordSubst = "keyword-subst"
)

type filterCommand struct {
	cmd *cobra.Command

	O struct {
		Kind     string
		Clean    bool
		Smudge   bool
		Filename string
	}
}

type filterDriver interface {
	Clean() error
	Smudge() error
}

func (v *filterCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "filter",
		Short: "Content filter drivers for git",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	v.cmd.Flags().StringVarP(&v.O.Kind,
		"kind",
		"k",
		filterKindKeywordSubst,
		"kind of filter driver")
	v.cmd.Flags().BoolVar(&v.O.Clean,
		"clean",
		false,
		"run clean command for filter")
	v.cmd.Flags().BoolVar(&v.O.Smudge,
		"smudge",
		false,
		"run smudge command for filter")

	return v.cmd
}

func (v filterCommand) Execute(args []string) error {
	var (
		err    error
		driver filterDriver
	)

	if !v.O.Clean && !v.O.Smudge {
		return fmt.Errorf("must provide one of --clean or --smudge option")
	}

	if v.O.Clean && v.O.Smudge {
		return fmt.Errorf("cannot use --clean and --smudge options together")
	}

	if len(args) != 1 {
		return fmt.Errorf("should provide one filename")
	}

	v.O.Filename = args[0]

	switch v.O.Kind {
	case filterKindKeywordSubst:
		driver = newKeywordSubstFilterDriver(v.O.Filename)
	default:
		return fmt.Errorf("known filter driver: %s", v.O.Kind)
	}

	if v.O.Clean {
		err = driver.Clean()
	} else {
		err = driver.Smudge()
	}

	return err
}

var filterCmd = filterCommand{}

func init() {
	rootCmd.AddCommand(filterCmd.Command())
}
