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

// Package path implements path related functions.
package path

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code.alibaba-inc.com/force/git-repo/errors"
)

// Macros for path
const (
	// DotRepo is ".repo".
	DotRepo = ".repo"
)

// HomeDir returns home directory.
func HomeDir() (string, error) {
	var (
		home string
	)

	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
	}
	if home == "" {
		home = os.Getenv("HOME")
	}

	if home == "" {
		return "", fmt.Errorf("cannot find HOME")
	}

	return home, nil
}

func xdgConfigHome(file string) (string, error) {
	var (
		home string
		err  error
	)

	home = os.Getenv("XDG_CONFIG_HOME")
	if home != "" {
		return filepath.Join(home, "git", file), nil
	}

	home, err = HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "git", file), nil
}

// ExpendHome expends path prefix "~/" to home dir.
func ExpendHome(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}

	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	if len(name) == 0 || name == "~" {
		return home, nil
	} else if len(name) > 1 && name[0] == '~' && (name[1] == '/' || name[1] == '\\') {
		return filepath.Join(home, name[2:]), nil
	}

	return filepath.Join(home, name), nil
}

// Abs returns absolute path and will expend homedir if path has "~/' prefix.
func Abs(name string) (string, error) {
	if name == "" {
		return os.Getwd()
	}

	if filepath.IsAbs(name) {
		return name, nil
	}

	if len(name) > 0 && name[0] == '~' && (len(name) == 1 || name[1] == '/' || name[1] == '\\') {
		return ExpendHome(name)
	}

	return filepath.Abs(name)
}

// AbsJoin returns absolute path, and use <dir> as parent dir for relative path.
func AbsJoin(dir, name string) (string, error) {
	if name == "" {
		return filepath.Abs(dir)
	}

	if filepath.IsAbs(name) {
		return name, nil
	}

	if len(name) > 0 && name[0] == '~' && (len(name) == 1 || name[1] == '/' || name[1] == '\\') {
		return ExpendHome(name)
	}

	return Abs(filepath.Join(dir, name))
}

// IsGitDir test whether dir is a valid git dir.
func IsGitDir(dir string) bool {
	var (
		err error
		fi  os.FileInfo
	)

	objectDir := filepath.Join(dir, "objects", "pack")
	if fi, err = os.Stat(objectDir); err != nil || !fi.IsDir() {
		return false
	}

	refsDir := filepath.Join(dir, "refs")
	if fi, err = os.Stat(refsDir); err != nil || !fi.IsDir() {
		return false
	}

	cfgFile := filepath.Join(dir, "config")
	if fi, err = os.Stat(cfgFile); err != nil || fi.IsDir() {
		return false
	}

	return true
}

// FindRepoRoot finds repo root path, where has a '.repo' subdir.
func FindRepoRoot(dir string) (string, error) {
	var (
		p   string
		err error
	)

	p, err = Abs(dir)
	if err != nil {
		return "", err
	}

	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return p, err
	}

	for {
		repodir := filepath.Join(p, DotRepo)
		if fi, err := os.Stat(repodir); err == nil {
			if fi.IsDir() {
				return p, nil
			}
			// We can use a .repo file to stop upward searching
			return "", errors.ErrRepoDirNotFound
		}

		oldP := p
		p = filepath.Dir(p)
		if oldP == p {
			// we reach the root dir
			break
		}
	}
	return "", errors.ErrRepoDirNotFound
}

// FindGitWorkSpace walks to upper directories to find gitdir and worktree.
func FindGitWorkSpace(dir string) (string, string, error) {
	var (
		err        error
		worktree   string
		repository string
	)

	dir, err = Abs(dir)
	if err != nil {
		return "", "", err
	}

	for {
		// Check if is in an repository
		if IsGitDir(dir) {
			repository = dir
			if filepath.Base(repository) == ".git" {
				worktree = filepath.Dir(repository)
			}
			break
		}

		// Check .git
		gitdir := filepath.Join(dir, ".git")
		fi, err := os.Stat(gitdir)
		if err != nil {
			// Test parent dir
			oldDir := dir
			dir = filepath.Dir(dir)
			if oldDir == dir {
				break
			}
			continue
		} else if fi.IsDir() {
			if IsGitDir(gitdir) {
				worktree = dir
				repository = gitdir
				break
			}
			return "", "", fmt.Errorf("corrupt git dir: %s", gitdir)
		} else {
			f, err := os.Open(gitdir)
			if err != nil {
				return "", "", fmt.Errorf("cannot open gitdir file '%s'", gitdir)
			}
			defer f.Close()
			reader := bufio.NewReader(f)
			line, err := reader.ReadString('\n')
			if strings.HasPrefix(line, "gitdir:") {
				realgit := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
				if !filepath.IsAbs(realgit) {
					realgit, err = AbsJoin(filepath.Dir(gitdir), realgit)
					if err != nil {
						return "", "", err
					}
				}
				if IsGitDir(realgit) {
					worktree = dir
					repository = realgit
					break
				}
				return "", "", fmt.Errorf("gitdir '%s' points to corrupt git repo: %s", gitdir, realgit)
			}
			return "", "", fmt.Errorf("bad gitdir file '%s'", gitdir)
		}
	}
	return worktree, repository, nil
}

// UnsetHome unsets HOME related environments.
func UnsetHome() {
	if runtime.GOOS == "windows" {
		os.Unsetenv("USERPROFILE")
		os.Unsetenv("HOMEDRIVE")
		os.Unsetenv("HOMEPATH")
	}
	os.Unsetenv("HOME")
}

// SetHome sets proper HOME environments.
func SetHome(home string) {
	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", home)
		if strings.Contains(home, ":\\") {
			slices := strings.SplitN(home, ":\\", 2)
			if len(slices) == 2 {
				os.Setenv("HOMEDRIVE", slices[0]+":")
				os.Setenv("HOMEPATH", "\\"+slices[1])
			}
		}
	} else {
		os.Setenv("HOME", home)
	}
}

// Exist check if path is exist.
func Exist(name string) bool {
	if _, err := os.Stat(name); err == nil {
		return true
	}
	return false
}
