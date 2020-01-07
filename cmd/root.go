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

// Package cmd implements verious git-repo subcommands using cobra as the
// command-line framework.
//
// Root command is defined in `cmd/root.go`, it implements public options
// for git-repo. Such as: --asume-yes, --quiet, --single.
//
// Each subcommand has a corresponding file, named `cmd/<subcmd>.go`, and
// the entrance of each subcommand is is defined in function `Execute(args)`.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/manifest"
	"github.com/aliyun/git-repo-go/version"

	log "github.com/jiangxin/multi-log"
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

// Response wraps error for subcommand, and is returned from cmd package.
type Response struct {
	// Err contains error returned from the subcommand executed.
	Err error

	// Cmd contains the command object.
	Cmd *cobra.Command
}

// IsUserError indicates it is a user fault, and should display the command
// usage in addition to displaying the error itself.
func (r Response) IsUserError() bool {
	return r.Err != nil && isUserError(r.Err)
}

type rootCommand struct {
	cmd *cobra.Command

	O struct {
		Version bool
	}
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
		// Do not want to show usage on every error
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	v.cmd.Flags().BoolVarP(&v.O.Version,
		"version",
		"V",
		false,
		"Show version")

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

func (v rootCommand) Execute(args []string) error {
	config.CheckGitAlias()
	if v.O.Version {
		showVersion()
	} else {
		return newUserError("run 'git repo -h' for help")
	}
	return nil
}

func (v rootCommand) checkGitVersion() {
	version.ValidateGitVersion()
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
		configDir, err := config.GetConfigDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filename := filepath.Join(configDir, config.DefaultGitRepoConfigFile)

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
