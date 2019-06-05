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
	"os"
	"path/filepath"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/versions"

	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	theRepoDir  string
	theWorkDir  string
	theManifest *manifest.Manifest
	rootCmd     = rootCommand{}
)

// The Response value from Execute.
type Response struct {
	// Err is set when the command failed to execute.
	Err error

	// The command that was executed.
	Cmd *cobra.Command
}

// IsUserError returns true is the Response error is a user error rather than a
// system error.
func (r Response) IsUserError() bool {
	return r.Err != nil && isUserError(r.Err)
}

type rootCommand struct {
	cmd *cobra.Command
}

// Command represents the base command when called without any subcommands
func (v *rootCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "git-repo",
		Short: "A command line tool for centralized git workflow",
		Long: `A command line tool for centralized git workflow.

Just like repo for the Android world, git-repo is a command line tool for
centralized git workflow of git core.

It can handle multiple repositories by using a manifest repository with
a default.xml file. And it can also handle a single repository by using
a '--single' opiton.

This tool is renamed as git-repo, so that wen can create git alias to run
this command with special options.`,
		Version: v.getVersion(),
		// Do not want to show usage on every error
		SilenceUsage: true,
	}

	v.cmd.PersistentFlags().StringVar(&cfgFile,
		"config",
		"",
		"config file (default is $HOME/.git-repo.yaml)")
	v.cmd.PersistentFlags().Bool("assume-no",
		false,
		"Automatic no to prompts")
	v.cmd.PersistentFlags().Bool("assume-yes",
		false,
		"Automatic yes to prompts")
	v.cmd.PersistentFlags().Bool("dryrun",
		false,
		"dryrun mode")
	v.cmd.PersistentFlags().BoolP("quiet",
		"q",
		false,
		"quiet mode")
	v.cmd.PersistentFlags().Bool("single",
		false,
		"single mode, no manifest")
	v.cmd.PersistentFlags().CountP("verbose",
		"v",
		"verbose mode")
	v.cmd.PersistentFlags().Bool("mock-no-symlink",
		false,
		"mock no symlink cap")
	v.cmd.PersistentFlags().Bool("mock-no-tty",
		false,
		"mock notty cap")
	v.cmd.PersistentFlags().String("mock-ssh-info-response",
		"",
		"mock remote ssh_info response")
	v.cmd.PersistentFlags().Int("mock-ssh-info-status",
		0,
		"mock remote ssh_info status")

	v.cmd.PersistentFlags().MarkHidden("assume-yes")
	v.cmd.PersistentFlags().MarkHidden("assume-no")
	v.cmd.PersistentFlags().MarkHidden("mock-ssh-info-status")
	v.cmd.PersistentFlags().MarkHidden("mock-ssh-info-response")
	v.cmd.PersistentFlags().MarkHidden("mock-no-symlink")
	v.cmd.PersistentFlags().MarkHidden("mock-no-tty")

	viper.BindPFlag(
		"assume-no",
		v.cmd.PersistentFlags().Lookup("assume-no"))
	viper.BindPFlag(
		"assume-yes",
		v.cmd.PersistentFlags().Lookup("assume-yes"))
	viper.BindPFlag(
		"dryrun",
		v.cmd.PersistentFlags().Lookup("dryrun"))
	viper.BindPFlag(
		"quiet",
		v.cmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag(
		"single",
		v.cmd.PersistentFlags().Lookup("single"))
	viper.BindPFlag(
		"verbose",
		v.cmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag(
		"mock-no-symlink",
		v.cmd.PersistentFlags().Lookup("mock-no-symlink"))
	viper.BindPFlag(
		"mock-no-tty",
		v.cmd.PersistentFlags().Lookup("mock-no-tty"))
	viper.BindPFlag(
		"mock-ssh-info-response",
		v.cmd.PersistentFlags().Lookup("mock-ssh-info-response"))
	viper.BindPFlag(
		"mock-ssh-info-status",
		v.cmd.PersistentFlags().Lookup("mock-ssh-info-status"))

	return v.cmd
}

// GetVersion is called by 'git repo --version'
func (v rootCommand) getVersion() string {
	config.InstallExtraGitConfig()
	return versions.GetVersion()
}

func (v rootCommand) checkGitVersion() {
	if !versions.ValidateGitVersion() {
		log.Fatalf("Please install or upgrade git to version %s or above",
			versions.MinGitVersion)
	}
}

func (v rootCommand) installConfigFiles() {
	var err error

	err = config.InstallExtraGitConfig()
	if err != nil {
		log.Error(err)
	}
	err = config.InstallRepoHooks()
	if err != nil {
		log.Error(err)
	}
	err = config.InstallRepoConfig()
	if err != nil {
		log.Error(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func (v rootCommand) initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		filename, err := path.Abs(config.DefaultGitRepoConfigFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".git-repo" (without extension).
		viper.AddConfigPath(filepath.Dir(filename))
		viper.SetConfigName(filepath.Base(filename))
	}

	viper.SetConfigType("yaml")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "ERROR: viper failed to read file %s: %s\n", viper.ConfigFileUsed(), err)
			os.Exit(1)
		}
	}
}

func (v rootCommand) initLog() {
	log.Init(log.Options{
		Verbose:       config.GetVerbose(),
		Quiet:         config.GetQuiet(),
		LogFile:       config.GetLogFile(),
		LogLevel:      config.GetLogLevel(),
		LogRotateSize: config.GetLogRotateSize(),
	})
}

func (v *rootCommand) AddCommand(cmds ...*cobra.Command) {
	v.Command().AddCommand(cmds...)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() Response {
	var (
		resp Response
	)

	c, err := rootCmd.Command().ExecuteC()
	resp.Err = err
	resp.Cmd = c
	return resp
}

func init() {
	cobra.OnInitialize(rootCmd.initConfig)
	cobra.OnInitialize(rootCmd.initLog)
	cobra.OnInitialize(rootCmd.checkGitVersion)
	cobra.OnInitialize(rootCmd.installConfigFiles)
}
