package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/errors"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
)

// WorkSpace is the toplevel structure for manipulating git-repo worktree
type WorkSpace struct {
	RootDir         string
	Manifest        *manifest.Manifest
	ManifestProject *project.ManifestProject
	Projects        []*project.Project
	projectByName   map[string][]*project.Project
	projectByPath   map[string]*project.Project
}

// Exists checks whether workspace is exist
func Exists(dir string) bool {
	manifestsDir := filepath.Join(dir, config.DotRepo, config.Manifests)
	if _, err := os.Stat(filepath.Join(manifestsDir, ".git")); err != nil {
		return false
	}
	cfg, err := goconfig.Load(manifestsDir)
	if err != nil {
		return false
	}
	return cfg.Get("remote.origin.url") != ""
}

// ManifestURL returns URL of manifest project
func (v *WorkSpace) ManifestURL() string {
	return v.Settings().ManifestURL
}

// Settings returns manifest project's Settings
func (v *WorkSpace) Settings() *project.RepoSettings {
	return v.ManifestProject.Settings
}

// Config returns git config file parser
func (v *WorkSpace) Config() goconfig.GitConfig {
	return v.ManifestProject.Config()
}

// SaveConfig will save config to git config file
func (v *WorkSpace) SaveConfig(cfg goconfig.GitConfig) error {
	return v.ManifestProject.SaveConfig(cfg)
}

// LinkManifest creates link of manifest.xml
func (v *WorkSpace) LinkManifest() error {
	if v.Settings().ManifestName != "" {
		if cap.Symlink() {
			target := filepath.Join(v.RootDir, config.DotRepo, config.ManifestXML)
			src, err := os.Readlink(target)
			if err != nil || filepath.Base(src) != v.Settings().ManifestName {
				os.Remove(target)
				src = filepath.Join(config.Manifests, v.Settings().ManifestName)
				log.Debugf("will symlink '%s' to '%s'", src, target)
				err = os.Symlink(src, target)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Load will read manifest XML file and reset ManifestURL if it changed,
// and reset URL of all projects in workspace.
func (v *WorkSpace) Load(manifestURL string) error {
	m, err := manifest.Load(filepath.Join(v.RootDir, config.DotRepo))
	if err != nil {
		return err
	}

	v.Manifest = m

	return v.loadProjects(manifestURL)
}

// Override will read alternate manifest XML file to initialize workspace
func (v *WorkSpace) Override(name string) error {
	manifestFile := filepath.Join(v.RootDir, config.DotRepo, config.Manifests, name)
	if _, err := os.Stat(manifestFile); err != nil {
		return fmt.Errorf("cannot find manifest: %s", name)
	}

	m, err := manifest.LoadFile(filepath.Join(v.RootDir, config.DotRepo), manifestFile)
	if err != nil {
		return err
	}
	v.Manifest = m

	return v.loadProjects("")
}

func (v *WorkSpace) loadProjects(manifestURL string) error {
	// Set manifest project even v.Manifest is nil
	v.ManifestProject = project.NewManifestProject(v.RootDir, manifestURL)

	// Set projects
	v.Projects = []*project.Project{}
	v.projectByName = make(map[string][]*project.Project)
	v.projectByPath = make(map[string]*project.Project)
	if v.Manifest != nil {
		for _, mp := range v.Manifest.AllProjects() {
			p := project.NewProject(&mp, v.ManifestProject.Settings)
			v.Projects = append(v.Projects, p)
			if _, ok := v.projectByName[p.Name]; !ok {
				v.projectByName[p.Name] = []*project.Project{}
			}
			v.projectByName[p.Name] = append(v.projectByName[p.Name], p)
			v.projectByPath[p.Path] = p
		}
	}

	return nil
}

// GetProjectsWithName returns projects which has matching name
func (v WorkSpace) GetProjectsWithName(name string) []*project.Project {
	return v.projectByName[name]
}

// GetProjectWithPath returns project which has matching path
func (v WorkSpace) GetProjectWithPath(p string) *project.Project {
	return v.projectByPath[p]
}

// GetProjectsOptions is options for GetProjects() function
type GetProjectsOptions struct {
	Groups       string
	MissingOK    bool
	SubmodulesOK bool
}

// GetProjects returns all matching projects
func (v WorkSpace) GetProjects(o *GetProjectsOptions, args ...string) ([]*project.Project, error) {
	var (
		groups      string
		result      = []*project.Project{}
		allProjects = []*project.Project{}
	)

	groups = o.Groups
	if groups == "" {
		groups = v.ManifestProject.Config().Get(config.CfgManifestGroups)
		if groups == "" {
			groups = "default,platform-" + runtime.GOOS
		}
	}

	if len(args) == 0 {
		allProjects = v.Projects
	} else {
		for _, arg := range args {
			ps := v.GetProjectsWithName(arg)
			if len(ps) == 0 {
				arg, _ = path.Abs(arg)
				absPath := filepath.ToSlash(arg)
				p := v.GetProjectWithPath(absPath)
				if p != nil {
					ps = append(ps, p)
				}
			}
			if len(ps) == 0 {
				return nil, errors.NoSuchProjectError(arg)
			}
			allProjects = append(allProjects, ps...)
		}
	}

	derivedProjects := []*project.Project{}
	for _, p := range allProjects {
		if o.SubmodulesOK || p.IsSyncS() {
			for _, sp := range p.GetSubmoduleProjects() {
				derivedProjects = append(derivedProjects, sp)
			}
		}
	}

	if len(derivedProjects) > 0 {
		allProjects = project.Join(allProjects, derivedProjects)
	}

	for _, p := range allProjects {
		if !o.MissingOK && !p.Exists() {
			if len(args) > 0 {
				return nil, errors.ProjectNoExistError(p.Name)
			}
			continue
		}
		if p.MatchGroups(groups) {
			result = append(result, p)
		} else if len(args) > 0 {
			return nil, errors.ProjectNotBelongToGroupsError(p.Name, groups)
		}
	}

	return result, nil
}

// NewWorkSpace finds and loads repo workspace. Will return an error if not found.
//
// 1. Searching a hidden `.repo` directory in `<dir>` or any parent directory.
// 2. Returns a WorkSpace objects based on the toplevel directory of workspace.
// 3. If cannot find valid repo workspace, return ErrRepoDirNotFound error.
func NewWorkSpace(dir string) (*WorkSpace, error) {
	var (
		err error
	)

	repoRoot, err := path.FindRepoRoot(dir)
	if err != nil {
		return nil, err
	}

	return newWorkSpace(repoRoot, "")
}

// NewWorkSpaceInit finds repo root and load specific manifest file.
// If workspace is not found, will use <dir> as root of a new workspace.
func NewWorkSpaceInit(dir, manifestURL string) (*WorkSpace, error) {
	var (
		err error
	)

	repoRoot, err := path.FindRepoRoot(dir)
	if err != nil {
		if err == errors.ErrRepoDirNotFound {
			repoRoot = dir
		} else {
			return nil, err
		}
	}

	return newWorkSpace(repoRoot, manifestURL)
}

func newWorkSpace(dir, manifestURL string) (*WorkSpace, error) {
	var (
		err error
	)

	if dir == "" {
		dir, err = path.Abs(dir)
		if err != nil {
			return nil, err
		}
	}

	ws := WorkSpace{RootDir: dir}
	err = ws.Load(manifestURL)
	if err != nil {
		return nil, err
	}

	return &ws, nil
}
