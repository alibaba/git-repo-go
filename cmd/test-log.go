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
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type testLogCommand struct {
	cmd *cobra.Command
}

func (v *testLogCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "log",
		Short: "test log",

		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	return v.cmd
}

func (v *testLogCommand) Execute(arts []string) error {
	log.WithField("my-key", "my-value").Trace("trace message, with fields...")
	log.WithFields(map[string]interface{}{"key1": "value1", "key2": "value2"}).Tracef("tracef message, with fields...")
	log.WithField("my-key", "my-value").Traceln("traceln message, with fields...")
	log.WithField("my-key", "my-value").Debug("debug message, with fields...")
	log.WithField("my-key", "my-value").Debugf("debugf message, with fields...")
	log.WithField("my-key", "my-value").Debugln("debugln message, with fields...")
	log.Debug("debug message...")
	log.Info("info message...")
	log.Warn("warn message...")
	log.Error("error message...")
	log.Notef("note message...")
	log.Printf("hello, %s.", "world")
	return nil
}

var testLogCmd = testLogCommand{}

func init() {
	testCmd.Command().AddCommand(testLogCmd.Command())
}
