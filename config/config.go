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

// Package config provides global variables, macros, environments and settings.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// DotRepo is '.repo', admin directory for git-repo
	DotRepo = path.DotRepo

	// CommitIDPattern indicates raw commit ID
	CommitIDPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

	// GitDefaultConfig is git global and system config.
	GitDefaultConfig goconfig.GitConfig
)

// Exported macros
const (
	GIT = "git"

	DefaultConfigPath = ".git-repo"
	DefaultLogRotate  = 20 * 1024 * 1024
	DefaultLogLevel   = "warn"

	CfgRepoArchive           = "repo.archive"
	CfgRepoDepth             = "repo.depth"
	CfgRepoDissociate        = "repo.dissociate"
	CfgRepoMirror            = "repo.mirror"
	CfgRepoReference         = "repo.reference"
	CfgRepoSubmodules        = "repo.submodules"
	CfgManifestGroups        = "manifest.groups"
	CfgManifestName          = "manifest.name"
	CfgRemoteOriginURL       = "remote.origin.url"
	CfgBranchDefaultMerge    = "branch.default.merge"
	CfgManifestRemoteType    = "manifest.remote.%s.type"
	CfgManifestRemoteVersion = "manifest.remote.%s.version"
	CfgManifestRemoteUser    = "manifest.remote.%s.user"
	CfgManifestRemoteSSHInfo = "manifest.remote.%s.sshinfo"
	CfgManifestRemoteExpire  = "manifest.remote.%s.expire"
	CfgAppGitRepoDisabled    = "app.git.repo.disabled"

	ManifestsDotGit  = "manifests.git"
	Manifests        = "manifests"
	DefaultXML       = "default.xml"
	ManifestXML      = "manifest.xml"
	LocalManifestXML = "local_manifest.xml"
	LocalManifests   = "local_manifests"
	ProjectObjects   = "project-objects"
	Projects         = "projects"

	ProtoTypeGerrit = "gerrit"
	ProtoTypeAGit   = "agit"

	RefsChanges = "refs/changes/"
	RefsMr      = "refs/merge-requests/"
	RefsHeads   = "refs/heads/"
	RefsTags    = "refs/tags/"
	RefsPub     = "refs/published/"
	RefsM       = "refs/remotes/m/"
	Refs        = "refs/"
	RefsRemotes = "refs/remotes/"

	MaxJobs = 32

	ViperEnvPrefix = "GIT_REPO"
)

// AssumeNo gets --asume-no option.
func AssumeNo() bool {
	return viper.GetBool("assume-no")
}

// AssumeYes gets --asume-yes option.
func AssumeYes() bool {
	return viper.GetBool("assume-yes")
}

// GetVerbose gets --verbose option.
func GetVerbose() int {
	return viper.GetInt("verbose")
}

// GetQuiet gets --quiet option.
func GetQuiet() bool {
	return viper.GetBool("quiet")
}

// IsSingleMode gets --single option.
func IsSingleMode() bool {
	return viper.GetBool("single")
}

// GetLogFile gets --logfile option.
func GetLogFile() string {
	logfile := viper.GetString("logfile")
	if logfile != "" && !filepath.IsAbs(logfile) {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logfile = filepath.Join(home, DefaultConfigPath, logfile)
	}
	return logfile
}

// GetLogLevel gets --loglevel option.
func GetLogLevel() string {
	return viper.GetString("loglevel")
}

// GetLogRotateSize gets logrotate size from config.
func GetLogRotateSize() int64 {
	logrotate := strings.ToLower(viper.GetString("logrotate"))
	if logrotate == "" {
		return 0
	}
	if logrotate[len(logrotate)-1] == 'b' {
		logrotate = logrotate[0 : len(logrotate)-1]
		if logrotate == "" {
			return 0
		}
	}
	scale := logrotate[len(logrotate)-1]
	if scale == 'k' || scale == 'm' || scale == 'g' {
		logrotate = logrotate[0 : len(logrotate)-1]
	}

	size, err := strconv.ParseInt(logrotate, 10, 64)
	if err != nil {
		log.Warnf("bad logrotate value: %s", viper.GetString("logrotate"))
		return 0
	}
	switch scale {
	case 'k':
		size <<= 10
	case 'm':
		size <<= 20
	case 'g':
		size <<= 30
	}
	return size
}

// NoCertChecks indicates whether ignore ssl cert.
func NoCertChecks() bool {
	return !GitDefaultConfig.GetBool("http.sslverify", true)
}

// GetMockSSHInfoStatus gets --mock-ssh-info-status option.
func GetMockSSHInfoStatus() int {
	return viper.GetInt("mock-ssh-info-status")
}

// GetMockSSHInfoResponse gets --mock-ssh-info-response option.
func GetMockSSHInfoResponse() string {
	return viper.GetString("mock-ssh-info-response")
}

// MockNoSymlink checks --mock-no-symlink option.
func MockNoSymlink() bool {
	return viper.GetBool("mock-no-symlink")
}

// MockNoTTY checks --mock-no-tty option.
func MockNoTTY() bool {
	return viper.GetBool("mock-no-tty")
}

// MockUploadOptionsEditScript gets --mock-upload-options-edit-script option.
func MockUploadOptionsEditScript() string {
	return viper.GetString("mock-upload-options-edit-script")
}

// IsDryRun gets --dryrun option.
func IsDryRun() bool {
	return viper.GetBool("dryrun")
}

func init() {
	viper.SetDefault("logrotate", DefaultLogRotate)
	viper.SetDefault("loglevel", DefaultLogLevel)

	viper.SetEnvPrefix(ViperEnvPrefix)
	viper.AutomaticEnv()

	GitDefaultConfig = goconfig.DefaultConfig()
}
