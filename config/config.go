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

package config

import (
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Define macros for config
const (
	DefaultConfigPath = ".git-repo"
	DefaultLogRotate  = 20 * 1024 * 1024
	DefaultLogLevel   = "warn"

	CfgRepoArchive    = "repo.archive"
	CfgRepoDepth      = "repo.depth"
	CfgRepoDissociate = "repo.dissociate"
	CfgRepoMirror     = "repo.mirror"
	CfgRepoReference  = "repo.reference"
	CfgRepoSubmodules = "repo.submodules"
	CfgManifestGroups = "manifest.groups"
	CfgManifestName   = "manifest.name"

	DotRepo          = ".repo"
	ManifestsDotGit  = "manifests.git"
	Manifests        = "manifests"
	DefaultXML       = "default.xml"
	ManifestXML      = "manifest.xml"
	LocalManifestXML = "local_manifest.xml"
	LocalManifests   = "local_manifests"
	ProjectObjects   = "project-objects"
	Projects         = "projects"

	ViperEnvPrefix = "GIT_REPO"
)

// GetVerbose gets --verbose option
func GetVerbose() int {
	return viper.GetInt("verbose")
}

// GetQuiet gets --quiet option
func GetQuiet() bool {
	return viper.GetBool("quiet")
}

// IsSingleMode checks --single option
func IsSingleMode() bool {
	return viper.GetBool("single")
}

// GetLogFile gets --logfile option
func GetLogFile() string {
	logfile := viper.GetString("logfile")
	if logfile != "" && !path.IsAbs(logfile) {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logfile = path.Join(home, DefaultConfigPath, logfile)
	}
	return logfile
}

// GetLogLevel gets --loglevel option
func GetLogLevel() string {
	return viper.GetString("loglevel")
}

// GetLogRotateSize gets logrotate size from config
func GetLogRotateSize() int64 {
	return viper.GetInt64("logrotate")
}

func init() {
	viper.SetDefault("logrotate", DefaultLogRotate)
	viper.SetDefault("loglevel", DefaultLogLevel)

	viper.SetEnvPrefix(ViperEnvPrefix)
	viper.AutomaticEnv()
}
