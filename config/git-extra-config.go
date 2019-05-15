package config

import (
	"os"
	"os/exec"
	"path/filepath"

	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
)

// Macros for git-extra-config
const (
	GitExtraConfigVersion = "2"
	GitExtraConfigFile    = "~/.git-repo/gitconfig"
	CfgRepoConfigVersion  = "repo.configversion"
)

func extraGitConfig() goconfig.GitConfig {
	cfg := goconfig.NewGitConfig()

	// Do not quote path, show UTF-8 characters directly
	cfg.Set("core.quotePath", false)

	cfg.Set("color.ui", "auto")

	// Add at most 20 commit logs in merge log message
	cfg.Set("merge.log", true)

	// Run git rebase with --autosquash option
	cfg.Set("rebase.autosquash", true)

	// Command alias
	cfg.Set("alias.br", "branch")
	cfg.Set("alias.ci", "commit -s")
	cfg.Set("alias.co", "checkout")
	cfg.Set("alias.cp", "cherry-pick")
	cfg.Set("alias.st", "status")
	cfg.Set("alias.logf", "log --pretty=fuller")
	cfg.Set("pretty.refs", "format:%h (%s, %ad)")
	cfg.Set("alias.logs", "log --pretty=refs  --date=short")

	// Alias commands for git-repo
	cfg.Set("alias.review", "repo upload --single")

	// Version
	cfg.Set(CfgRepoConfigVersion, GitExtraConfigVersion)

	return cfg
}

func saveExtraGitConfig() error {
	var (
		err error
	)

	filename, _ := path.Abs(GitExtraConfigFile)
	dir := filepath.Dir(filename)

	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	cfg := extraGitConfig()
	err = cfg.Save(filename)
	return err
}

// InstallExtraGitConfig if necessary
func InstallExtraGitConfig() error {
	var err error

	globalConfig, err := goconfig.GlobalConfig()
	version := globalConfig.Get(CfgRepoConfigVersion)
	if version == GitExtraConfigVersion {
		return nil
	}

	log.Debugf("unmatched git config version: %s != %s", version, GitExtraConfigVersion)
	found := false
	gitExtraConfigFile, _ := path.Abs(GitExtraConfigFile)
	for _, p := range globalConfig.GetAll("include.path") {
		p, _ = path.Abs(p)
		if p == gitExtraConfigFile {
			found = true
			break
		}
	}
	if !found {
		cmds := []string{"git",
			"config",
			"--global",
			"--add",
			"include.path",
			GitExtraConfigFile,
		}
		err = exec.Command(cmds[0], cmds[1:]...).Run()
		if err != nil {
			return err
		}
	}

	err = saveExtraGitConfig()
	if err != nil {
		return err
	}
	return nil
}
