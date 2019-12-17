package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/path"
)

const (
	// gitHooksVersion is version of hooks, and should be promoted if
	// anything changed in hooks
	gitHooksVersion = "1"

	// gerritCommitMsgHook is content of commit-msg hook for Gerrit
	gerritCommitMsgHook = `#!/bin/sh
# From Gerrit Code Review 3.0.0-rc3-236-g33e7081a25
#
# Part of Gerrit Code Review (https://www.gerritcodereview.com/)
#
# Copyright (C) 2009 The Android Open Source Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# avoid [[ which is not POSIX sh.
if test "$#" != 1 ; then
  echo "$0 requires an argument."
  exit 1
fi

if test ! -f "$1" ; then
  echo "file does not exist: $1"
  exit 1
fi

# Do not create a change id if requested
if test "false" = "$(git config --bool --get gerrit.createChangeId)" ; then
  exit 0
fi

# $RANDOM will be undefined if not using bash, so don't use set -u
random=$( (whoami ; hostname ; date; cat "$1" ; echo $RANDOM) | git hash-object --stdin)
dest="$1.tmp.${random}"

trap 'rm -f "${dest}"' EXIT

if ! git stripspace --strip-comments < "$1" > "${dest}" ; then
   echo "cannot strip comments from $1"
   exit 1
fi

if test ! -s "${dest}" ; then
  echo "file is empty: $1"
  exit 1
fi

# Avoid the --in-place option which only appeared in Git 2.8
# Avoid the --if-exists option which only appeared in Git 2.15
if ! git -c trailer.ifexists=doNothing interpret-trailers \
      --trailer "Change-Id: I${random}" < "$1" > "${dest}" ; then
  if ! grep -q "^ChangeId: I" "${dest}"; then
    echo "Change-Id: I${random}" >>"${dest}"
  fi
fi

if ! mv "${dest}" "$1" ; then
  echo "cannot mv ${dest} to $1"
  exit 1
fi
`
)

var (
	// GerritHooks defines map of Gerrit hooks.
	GerritHooks = map[string]string{
		"commit-msg": gerritCommitMsgHook,
	}
)

// GetRepoHooksDir returns the hooks template dir for git-repo.
func GetRepoHooksDir() (string, error) {
	cfgDir, err := GetConfigDir()
	if err != nil {
		return "", fmt.Errorf("fail to get hooks dir: %s", err)
	}
	return filepath.Join(cfgDir, "hooks"), nil
}

func hooksVersionFile() string {
	hooksDir, _ := GetRepoHooksDir()
	return filepath.Join(hooksDir, "VERSION")
}

func isHooksUptodate() bool {
	if !path.Exist(hooksVersionFile()) {
		return false
	}
	data, err := ioutil.ReadFile(hooksVersionFile())
	if err != nil {
		return false
	}
	if strings.TrimSpace(string(data)) != gitHooksVersion {
		return false
	}
	return true
}

// InstallRepoHooks installs hooks into ~/.git-repo/hooks.
func InstallRepoHooks() error {
	if isHooksUptodate() {
		return nil
	}

	hooksDir, err := GetRepoHooksDir()
	if err != nil {
		return err
	}
	if !path.Exist(hooksDir) {
		err = os.MkdirAll(hooksDir, 0755)
		if err != nil {
			return fmt.Errorf("fail to install hooks: %s", err)
		}
	}
	for name, data := range GerritHooks {
		file := filepath.Join(hooksDir, name)
		finfo, err := os.Stat(file)
		if err != nil || int(finfo.Size()) != len(data) {
			err = ioutil.WriteFile(file, []byte(data), 0755)
			if err != nil {
				return fmt.Errorf("fail to write hooks: %s", err)
			}
		}
	}
	err = ioutil.WriteFile(hooksVersionFile(), []byte(gitHooksVersion+"\n"), 0644)
	return nil
}
