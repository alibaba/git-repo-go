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
	"code.alibaba-inc.com/force/git-repo/version"

	"github.com/jiangxin/multi-log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	theRepoDir string
	theWorkDir string
)

// Define macros for git-repo
const (
	DefaultConfigFile = ".git-repo"
	EnvPrefix         = "GIT_REPO"
	RepoDir           = ".repo"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-repo",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Version: version.GetVersion(),
	// Do not want to show usage on every error
	SilenceUsage:     true,
	PersistentPreRun: findRepo,
}

// find .repo dir
func findRepo(cmd *cobra.Command, args []string) {
	if config.IsSingleMode() {
		findRepoSingle(cmd, args)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("cannot get current dir")
	}
	p, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		log.Warnf("fail to call EvalSymlinks on %s", cwd)
	}

	for {
		repoDir := filepath.Join(p, RepoDir)
		if fi, err := os.Stat(repoDir); err == nil && fi.IsDir() {
			theRepoDir = repoDir
			theWorkDir = p
			break
		}

		oldP := p
		p = filepath.Dir(p)
		if oldP == p {
			// we reach the root dir
			break
		}
	}
}

// find current repo rootdir
func findRepoSingle(cmd *cobra.Command, args []string) {
	// TODO: find git dir and worktree
}

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() Response {
	var resp Response

	c, err := rootCmd.ExecuteC()
	resp.Err = err
	resp.Cmd = c
	return resp
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initLog)
	cobra.OnInitialize(checkVersion)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-repo.yaml)")
	rootCmd.PersistentFlags().CountP("verbose", "v", "verbose mode")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().Bool("single", false, "single mode, no manifest")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("single", rootCmd.PersistentFlags().Lookup("single"))
}

func checkVersion() {
	if !version.ValidateGitVersion() {
		log.Fatalf("Please install or upgrade git to version %s or above",
			version.MinGitVersion)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".git-repo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(DefaultConfigFile)
	}

	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "ERROR: viper failed to read file %s: %s\n", viper.ConfigFileUsed(), err)
			os.Exit(1)
		}
	}
}

func initLog() {
	log.Init(log.Options{
		Verbose:       config.GetVerbose(),
		Quiet:         config.GetQuiet(),
		LogFile:       config.GetLogFile(),
		LogLevel:      config.GetLogLevel(),
		LogRotateSize: config.GetLogRotate(),
	})
}
