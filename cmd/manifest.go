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
	"io"
	"os"

	"code.alibaba-inc.com/force/git-repo/manifest"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type manifestCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   struct {
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

func (v manifestCommand) WriteManifest(writer io.Writer) error {
	ws := v.RepoWorkSpace()

	if v.O.PegRev {
		err := ws.FreezeManifest(!v.O.PegRevNoUpstream)
		if err != nil {
			return err
		}
	}

	data, err := manifest.Marshal(ws.Manifest)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	if data[len(data)-1] != '\n' {
		writer.Write([]byte("\n"))
	}

	if v.O.OutputFile != "-" {
		log.Notef("Saved manifest to %s", v.O.OutputFile)
	}
	return nil
}

func (v manifestCommand) Execute(args []string) error {
	var (
		writer io.ReadWriteCloser
	)

	if v.O.OutputFile == "" {
		log.Fatal("no output file, no operation to perform")
	} else if v.O.OutputFile == "-" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(v.O.OutputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		writer = file
		defer file.Close()
	}

	return v.WriteManifest(writer)
}

var manifestCmd = manifestCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: true,
		SingleOK: false,
	},
}

func init() {
	rootCmd.AddCommand(manifestCmd.Command())
}
