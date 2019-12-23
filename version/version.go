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

// Package version implements versions related functions
package version

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/jiangxin/multi-log"
)

var (
	// Version is the verison of git-repo.
	Version = "undefined"

	// GitVersion is version of git.
	GitVersion = ""
)

type gitCompatibleIssue struct {
	Version string
	Message string
	Fatal   bool
}

var gitCompatibleIssues = []gitCompatibleIssue{
	gitCompatibleIssue{
		"1.7.10",
		"Git config extension (by include.path directive) is only supported in git 1.7.10 and above",
		true,
	},
	gitCompatibleIssue{
		"2.2.0",
		"The git-interpret-trailers command introduced in git 2.2.0 is used for the Gerrit commit-msg hook.",
		false,
	},
	gitCompatibleIssue{
		"2.9.0",
		"sending custom HTTP headers is only supported in git 2.9.0 and above",
		false,
	},
	gitCompatibleIssue{
		"2.10.0",
		"Some review options are sent using git push-options which are available in git 2.10.0 and above.",
		false,
	},
}

// GetVersion returns git-repo version.
func GetVersion() string {
	return Version
}

// GetGitVersion gets current installed git version.
func GetGitVersion() string {
	var out bytes.Buffer

	cmd := exec.Command("git", "version")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("fail to run git version")
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(out.String(), "git version "))
}

// CompareVersion compares two versions.
func CompareVersion(_left, _right string) int {
	left := strings.Split(_left, ".")
	right := strings.Split(_right, ".")
	pos := len(left)
	if pos > len(right) {
		pos = len(right)
	}

	for i := 0; i < pos; i++ {
		l, lErr := strconv.Atoi(left[i])
		r, rErr := strconv.Atoi(right[i])

		if lErr != nil && rErr != nil {
			if left[i] > right[i] {
				return 1
			} else if left[i] < right[i] {
				return -1
			} else {
				continue
			}
		} else if lErr != nil {
			return -1
		} else if rErr != nil {
			return 1
		}

		if l > r {
			return 1
		} else if l < r {
			return -1
		}
	}

	if len(left) > len(right) {
		if _, err := strconv.Atoi(left[pos]); err == nil {
			return 1
		}
		return -1
	} else if len(left) < len(right) {
		if _, err := strconv.Atoi(right[pos]); err == nil {
			return -1
		}
		return 1
	}

	return 0
}

// ValidateGitVersion is used to check installed git version.
func ValidateGitVersion() {
	var (
		// lower conflict verison
		lcVersion string
		// higher compatible version
		hcVersion      string
		messages       []string
		suppressIssues bool
	)

	if _, ok := os.LookupEnv("GIT_REPO_SUPPRESS_COMPATIBLE_ISSUES"); ok {
		suppressIssues = true
	}

	for _, issue := range gitCompatibleIssues {
		if CompareVersion(GitVersion, issue.Version) < 0 {
			if issue.Fatal {
				if CompareVersion(issue.Version, lcVersion) > 0 {
					lcVersion = issue.Version
				}
			} else {
				if CompareVersion(issue.Version, hcVersion) > 0 {
					hcVersion = issue.Version
				}
			}
			messages = append(messages, issue.Message)
		}
	}

	if GitVersion == "" {
		log.Errorf("Please install git to version %s or above", lcVersion)
	} else if lcVersion != "" {
		log.Errorf("Please upgrade git to version %s or above", lcVersion)
	} else if hcVersion != "" {
		if !suppressIssues {
			log.Warnf("You are suggested to install or upgrade git to version %s or above",
				hcVersion)
		}
	} else {
		return
	}

	if !suppressIssues {
		for _, msg := range messages {
			fmt.Printf("\t* %s\n", msg)
		}
	}

	if lcVersion != "" || GitVersion == "" {
		os.Exit(1)
	}
}

func init() {
	GitVersion = GetGitVersion()
}
