package path

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"code.alibaba-inc.com/force/git-repo/config"
	"github.com/jiangxin/multi-log"
)

func xdgConfigHome(file string) string {
	home := os.Getenv("XDG_CONFIG_HOME")
	if home != "" {
		return filepath.Join(home, "git", file)
	}

	home = homeDir()
	if home != "" {
		return filepath.Join(home, ".config", "git", file)
	}
	return ""
}

func homeDir() string {
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
		fmt.Fprintln(os.Stderr, "ERROR: fail to get HOME")
	}

	return home
}

func expendHome(file string) string {
	if filepath.IsAbs(file) {
		return file
	}

	home := homeDir()
	if home == "" {
		return file
	}

	if len(file) == 0 {
		return home
	}

	if file[0] == '~' {
		if len(file) == 1 {
			return home
		} else if file[1] == '/' || file[1] == '\\' {
			return filepath.Join(home, file[2:])
		}
	}

	return filepath.Join(home, file)
}

// Abs makes a absolute path and will resolve ~/
func Abs(name string) string {
	if filepath.IsAbs(name) {
		return name
	}

	if len(name) > 0 && name[0] == '~' && (len(name) == 1 || name[1] == '/' || name[1] == '\\') {
		return expendHome(name)
	}

	name, _ = filepath.Abs(name)
	return name
}

// AbsJoin joins path to dir instead of cwd, and makes it an absolute path
func AbsJoin(dir, name string) string {
	if filepath.IsAbs(name) {
		return name
	}

	if len(name) > 0 && name[0] == '~' && (len(name) == 1 || name[1] == '/' || name[1] == '\\') {
		return expendHome(name)
	}

	return Abs(filepath.Join(dir, name))
}

// FindRepoRoot finds repo root path, where has a '.repo' subdir
func FindRepoRoot(dir string) string {
	var (
		err error
	)

	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatal("cannot get current dir")
		}
	}
	p, err := filepath.EvalSymlinks(dir)
	if err != nil {
		log.Warnf("fail to call EvalSymlinks on %s", dir)
		p = dir
	}

	for {
		repoDir := filepath.Join(p, config.RepoDir)
		if fi, err := os.Stat(repoDir); err == nil && fi.IsDir() {
			return p
		}

		oldP := p
		p = filepath.Dir(p)
		if oldP == p {
			// we reach the root dir
			break
		}
	}
	return ""
}
