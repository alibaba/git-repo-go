package workspace

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/alibaba/git-repo-go/cap"
	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/errors"
	"github.com/alibaba/git-repo-go/file"
	"github.com/alibaba/git-repo-go/manifest"
	"github.com/alibaba/git-repo-go/path"
	"github.com/alibaba/git-repo-go/project"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
)

// RepoWorkSpace is the toplevel structure for manipulating git-repo worktree.
type RepoWorkSpace struct {
	RootDir         string
	Manifest        *manifest.Manifest
	ManifestProject *project.ManifestProject
	Projects        []*project.Project
	projectByName   map[string][]*project.Project
	projectByPath   map[string]*project.Project
	httpClient      *http.Client
}

// Exists checks whether workspace is exist.
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

// AdminDir returns .repo dir.
func (v RepoWorkSpace) AdminDir() string {
	return filepath.Join(v.RootDir, config.DotRepo)
}

// IsSingle is false for workspace initialized by manifests project.
func (v RepoWorkSpace) IsSingle() bool {
	return false
}

// IsMirror indicates whether repo is in mirror mode
func (v RepoWorkSpace) IsMirror() bool {
	return v.Settings().Mirror
}

// ManifestURL returns URL of manifest project.
func (v *RepoWorkSpace) ManifestURL() string {
	return v.Settings().ManifestURL
}

// Settings returns manifest project's Settings.
func (v *RepoWorkSpace) Settings() *project.RepoSettings {
	return v.ManifestProject.Settings
}

// Config returns git config file parser.
func (v *RepoWorkSpace) Config() goconfig.GitConfig {
	return v.ManifestProject.Config()
}

// SaveConfig will save config to git config file.
func (v *RepoWorkSpace) SaveConfig(cfg goconfig.GitConfig) error {
	return v.ManifestProject.SaveConfig(cfg)
}

// LinkManifest creates link of manifest.xml.
func (v *RepoWorkSpace) LinkManifest() error {
	srcAbs := filepath.Join(v.RootDir, config.DotRepo, config.Manifests, v.Settings().ManifestName)
	srcRel := filepath.Join(config.Manifests, v.Settings().ManifestName)

	if !path.Exist(srcAbs) {
		return fmt.Errorf("link manifest failed, cannot find file '%s'", srcRel)
	}
	if v.Settings().ManifestName != "" {
		if cap.CanSymlink() {
			target := filepath.Join(v.RootDir, config.DotRepo, config.ManifestXML)
			linkedSrc, err := os.Readlink(target)
			if err != nil || filepath.Base(linkedSrc) != v.Settings().ManifestName {
				os.Remove(target)
				log.Debugf("will symlink '%s' to '%s'", srcRel, target)
				err = os.Symlink(srcRel, target)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Load will read manifest XML file and reset ManifestURL if it changed
// and reset URL of all projects in workspace.
func (v *RepoWorkSpace) load(manifestURL string) error {
	m, err := manifest.Load(filepath.Join(v.RootDir, config.DotRepo))
	if err != nil {
		return err
	}

	v.Manifest = m

	return v.loadProjects(manifestURL)
}

// Override will read alternate manifest XML file to initialize workspace.
func (v *RepoWorkSpace) Override(name string) error {
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

func (v *RepoWorkSpace) manifestsProjectName() string {
	if v.Manifest == nil {
		return "manifests"
	}

	u := v.ManifestProject.ManifestURL()
	fetch := ""
	for _, r := range v.Manifest.Remotes {
		if r.Fetch != "" && r.Fetch[0] == '.' {
			if len(fetch) < len(r.Fetch) {
				fetch = r.Fetch
			}
		}
	}
	return manifestsProjectName(u, fetch)
}

func manifestsProjectName(url, fetch string) string {
	if strings.Contains(url, "://") {
		url = strings.SplitN(url, "://", 2)[1]
	}
	if strings.HasSuffix(url, ".git") {
		url = strings.TrimSuffix(url, ".git")
	}

	if url == "" || fetch == "" {
		return "manifests"
	}

	level := 1
	for _, dir := range strings.Split(fetch, "/") {
		if dir == ".." {
			level++
		} else if dir != "" && dir != "." {
			level--
		}
	}

	dirs := strings.Split(url, "/")
	if level > len(dirs) {
		level = len(dirs)
	}
	if level == 0 {
		return "manifests"
	}

	return filepath.Join(dirs[len(dirs)-level:]...)
}

func (v *RepoWorkSpace) loadProjects(manifestURL string) error {
	var p *project.Project
	// Set manifest project even v.Manifest is nil
	v.ManifestProject = project.NewManifestProject(v.RootDir, manifestURL)
	s := v.ManifestProject.Settings

	// Check whether git-repo is disabled.
	if v.ManifestProject.Config().GetBool(config.CfgAppGitRepoDisabled, false) {
		return fmt.Errorf("git-repo is disabled for this workspace, use repo instead")
	}

	// Set projects
	v.Projects = []*project.Project{}
	v.projectByName = make(map[string][]*project.Project)
	v.projectByPath = make(map[string]*project.Project)

	if v.Manifest != nil {
		allProjects := v.Manifest.AllProjects()
		if s.Mirror {
			mp := *manifest.ManifestsProject
			mp.Name = v.manifestsProjectName()
			allProjects = append(allProjects, mp)
		}

		for _, mp := range allProjects {
			if s.Mirror {
				p = project.NewMirrorProject(&mp, v.ManifestProject.Settings, v.Manifest)
				// Only save one of projects with the same name
				if _, ok := v.projectByName[p.Name]; ok {
					continue
				}
			} else {
				p = project.NewProject(&mp, v.ManifestProject.Settings, v.Manifest)
			}
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

// GetProjectsWithName returns projects which has matching name.
func (v RepoWorkSpace) GetProjectsWithName(name string) []*project.Project {
	return v.projectByName[name]
}

// GetProjectWithPath returns project which has matching path.
func (v RepoWorkSpace) GetProjectWithPath(p string) *project.Project {
	return v.projectByPath[p]
}

// GetProjectsOptions is options for GetProjects() function.
type GetProjectsOptions struct {
	Groups       string
	MissingOK    bool
	SubmodulesOK bool
}

// GetProjects returns all matching projects.
func (v RepoWorkSpace) GetProjects(o *GetProjectsOptions, args ...string) ([]*project.Project, error) {
	var (
		groups      string
		result      = []*project.Project{}
		allProjects = []*project.Project{}
		pDir        string
	)

	cwd, _ := os.Getwd()
	cwd, _ = filepath.EvalSymlinks(cwd)
	pDir, _ = filepath.Rel(v.RootDir, cwd)

	if o == nil {
		o = &GetProjectsOptions{}
	}
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
				if pDir != "" {
					arg = filepath.Clean(filepath.Join(pDir, arg))
				}
				p := v.GetProjectWithPath(arg)
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

type freezeProject struct {
	WorkSpace    *RepoWorkSpace
	FillUpstream bool
}

func (v *freezeProject) Process(mp *manifest.Project, parentDir string) error {
	var (
		rev string
		err error
	)

	if parentDir == "" {
		parentDir = mp.Path
	} else {
		parentDir = filepath.Join(parentDir, mp.Path)
	}

	p := v.WorkSpace.GetProjectWithPath(parentDir)
	if p == nil {
		log.Warnf("cannot find project '%s' to freeze", parentDir)
		return nil
	}

	if v.WorkSpace.Settings().Mirror {
		rev, err = p.ResolveRevision(p.Revision)
		if err != nil {
			log.Warn(err)
			return nil
		}
	} else {
		rev, err = p.ResolveRevision("HEAD")
		if err != nil {
			log.Warn(err)
			return nil
		}
	}

	if v.FillUpstream {
		if p.Upstream != "" {
			mp.Upstream = p.Upstream
		} else {
			mp.Upstream = p.Revision
		}
	}

	mp.Revision = rev
	return nil
}

// FreezeManifest changes projects of manifest, and set revision of project to
// fixed revision.
func (v *RepoWorkSpace) FreezeManifest(fillUpstream bool) error {

	handle := &freezeProject{
		WorkSpace:    v,
		FillUpstream: fillUpstream,
	}
	return v.Manifest.ProjectHandle(handle)
}

// UpdateProjectList updates `project.list` file and try to remove obsolete projects.
func (v *RepoWorkSpace) UpdateProjectList(submodulesOK bool) ([]string, error) {
	var (
		newPaths = []string{}
		oldPaths = []string{}
	)

	allProjects, err := v.GetProjects(&GetProjectsOptions{
		MissingOK:    true,
		SubmodulesOK: submodulesOK,
	})
	if err != nil {
		return nil, err
	}

	for _, p := range allProjects {
		newPaths = append(newPaths, p.Path)
	}

	projectListFile := filepath.Join(v.RootDir, config.DotRepo, "project.list")
	if _, err = os.Stat(projectListFile); err == nil {
		f, err := os.Open(projectListFile)
		if err != nil {
			log.Fatalf("fail to open %s: %s", projectListFile, err)
		}
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString('\n')
			line = strings.TrimSpace(line)
			if line != "" {
				oldPaths = append(oldPaths, line)
			}
			if err != nil {
				break
			}
		}
		f.Close()
	}

	remains, err := v.removeObsoletePaths(oldPaths, newPaths)
	if err != nil {
		log.Errorln(err)
	}

	projectListLockFile := projectListFile + ".lock"
	lockf, err := file.New(projectListLockFile).OpenCreateRewriteExcl()
	if err != nil {
		return nil, fmt.Errorf("fail to create lockfile '%s': %s", projectListLockFile, err)
	}
	defer lockf.Close()
	for _, p := range newPaths {
		_, err = lockf.WriteString(p + "\n")
		if err != nil {
			return nil, fmt.Errorf("fail to save lockfile '%s': %s", projectListLockFile, err)
		}
	}
	for _, p := range remains {
		_, err = lockf.WriteString(p + "\n")
		if err != nil {
			return nil, fmt.Errorf("fail to save lockfile '%s': %s", projectListLockFile, err)
		}
	}
	lockf.Close()

	err = os.Rename(projectListLockFile, projectListFile)
	if err != nil {
		return nil, fmt.Errorf("fail to rename lockfile to '%s': %s", projectListFile, err)
	}

	return remains, nil
}

// ShortGitObjectsDir is relative path of shared objects gitdir
func (v *RepoWorkSpace) ShortGitObjectsDir(name string) string {
	if name == "" {
		return ""
	}
	if v.IsMirror() {
		return name + ".git"
	}
	return filepath.Join(
		config.DotRepo,
		config.ProjectObjects,
		name+".git",
	)
}

// ShortGitDir is relative path of gitdir
func (v *RepoWorkSpace) ShortGitDir(pathName string) string {
	if pathName == "" {
		return ""
	}
	if v.IsMirror() {
		return ""
	}
	return filepath.Join(
		config.DotRepo,
		config.Projects,
		pathName+".git",
	)
}

func (v *RepoWorkSpace) removeObsoletePaths(oldPaths, newPaths []string) ([]string, error) {
	var (
		remains []string = []string{}
		errMsgs []string = []string{}
		err     error
	)

	obsoletePaths := findObsoletePaths(oldPaths, newPaths)

	for _, obsoletePath := range obsoletePaths {
		var workdirAbs, gitdir, gitdirAbs string

		p := filepath.Clean(obsoletePath)
		if p == "" || p == "." || p == ".." ||
			strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") {
			errMsgs = append(errMsgs, fmt.Sprintf("bad project path '%s', ignored", obsoletePath))
			continue
		}
		workdirAbs = filepath.Clean(filepath.Join(v.RootDir, p))
		gitdir = v.ShortGitDir(p)
		if gitdir != "" {
			gitdirAbs = filepath.Clean(filepath.Join(v.RootDir, gitdir))
		}
		/*
			workRepoPath := filepath.Clean(filepath.Join(v.RootDir,
				config.DotRepo,
				config.Projects,
				p+".git"))
		*/

		if !strings.HasPrefix(workdirAbs, v.RootDir) {
			// Ignore bad path, and do not update remains.
			errMsgs = append(errMsgs,
				fmt.Sprintf("project '%s', which beyond repo root '%s', ignored",
					p, v.RootDir))
			continue
		}
		if strings.TrimPrefix(workdirAbs, v.RootDir) == "" {
			continue
		}

		// Obsolete path does not exist, ignore it.
		if !path.Exist(workdirAbs) {
			// Remove gitdirAbs if exists
			if gitdirAbs != "" && path.Exist(gitdirAbs) {
				err = os.RemoveAll(gitdirAbs)
				if err != nil {
					remains = append(remains, p)
					errMsgs = append(errMsgs, fmt.Sprintf("fail to remove '%s': %s", gitdir, err))
				} else {
					v.removeEmptyDirs(gitdirAbs)
				}
			}
			continue
		}

		// workdirAbs is not a git workdir
		if !path.Exist(filepath.Join(workdirAbs, ".git")) {
			remains = append(remains, p)
			errMsgs = append(errMsgs,
				fmt.Sprintf("obsolete path '%s' does not seem like a git worktree", p))
			continue
		}

		if ok, _ := project.IsClean(workdirAbs); !ok {
			remains = append(remains, p)
			errMsgs = append(errMsgs,
				fmt.Sprintf(`cannot remove project "%s": uncommitted changes are present.
Please commit changes, upload then run sync again`,
					p))
			continue
		}

		/*
		 * TODO: Do not remove project p, if:
		 * - There are commits in branches which are not in upstream.
		 * - There are stash file, which may have changes not be committed.
		 */

		// TODO: Before making a implementation to check precious in obsolete path,
		// not remove projects, and leave them for users to delete.
		remains = append(remains, p)
		continue

		/*
			// Remove worktree, except recursive git repositories
			ignoreRepos := findGitWorktree(workdir)
			err = removeWorktree(workdir, ignoreRepos)
			if err != nil {
				remains = append(remains, p)
				errMsgs = append(errMsgs, fmt.Sprintf("fail to remove '%s': %s", workdir, err))
			} else {
				v.removeEmptyDirs(workdir)
			}
		*/
	}

	if len(errMsgs) > 0 {
		err = fmt.Errorf(strings.Join(errMsgs, "\n"))
	}

	return remains, err
}

// findObsoletePaths returns obsolete paths.
func findObsoletePaths(oldPaths, newPaths []string) []string {
	result := []string{}
	i, j := 0, 0

	// oldPaths and newPaths must be sorted.
	sort.Strings(oldPaths)
	sort.Strings(newPaths)

	for i < len(oldPaths) && j < len(newPaths) {
		if oldPaths[i] < newPaths[j] {
			result = append(result, oldPaths[i])
			i++
		} else if oldPaths[i] > newPaths[j] {
			j++
		} else {
			i++
			j++
		}
	}
	for i < len(oldPaths) {
		result = append(result, oldPaths[i])
		i++
	}

	return result
}

func removeWorktree(dir string, gitTrees []string) error {
	var err error

	if len(gitTrees) == 0 {
		log.Printf("will remove %s", dir)
		return os.RemoveAll(dir)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Printf("will remove %s", path)
			err = os.Remove(path)
			if err != nil {
				return err
			}
			return nil
		}
		for _, p := range gitTrees {
			if path == p {
				return filepath.SkipDir
			}
			if strings.HasPrefix(p, path+string(os.PathSeparator)) || strings.HasPrefix(p, path+"/") {
				return nil
			}
		}
		log.Printf("will remove all %s", path)
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		return filepath.SkipDir
	})

	return err
}

func findGitWorktree(dir string) []string {
	result := []string{}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, err = os.Stat(filepath.Join(path, ".git")); err == nil {
				result = append(result, path)
				return filepath.SkipDir
			}
		}
		return nil
	})

	return result
}

func (v *RepoWorkSpace) removeEmptyDirs(dir string) {
	var (
		oldDir = dir
	)

	for {
		if !strings.HasPrefix(dir, v.RootDir) {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
		if dir == oldDir {
			break
		}
		oldDir = dir
	}
}

// NewRepoWorkSpace finds and loads repo workspace. Will return an error if not found.
//
// 1. Searching a hidden `.repo` directory in `<dir>` or any parent directory.
// 2. Returns a RepoWorkSpace objects based on the toplevel directory of workspace.
// 3. If cannot find valid repo workspace, return ErrRepoDirNotFound error.
func NewRepoWorkSpace(dir string) (*RepoWorkSpace, error) {
	var (
		err error
	)

	topDir, err := path.FindTopDir(dir)
	if err != nil {
		return nil, err
	}

	return newRepoWorkSpace(topDir, "")
}

// NewEmptyRepoWorkSpace returns empty workspace for new created workspace.
func NewEmptyRepoWorkSpace(dir, manifestURL string) (*RepoWorkSpace, error) {
	var (
		err error
	)

	if dir == "" {
		dir, err = path.Abs(dir)
		if err != nil {
			return nil, err
		}
	}

	ws := RepoWorkSpace{RootDir: dir}
	ws.Manifest = nil
	err = ws.loadProjects(manifestURL)

	if err != nil {
		return nil, err
	}

	return &ws, nil
}

func newRepoWorkSpace(dir, manifestURL string) (*RepoWorkSpace, error) {
	var (
		err error
	)

	if dir == "" {
		dir, err = path.Abs(dir)
		if err != nil {
			return nil, err
		}
	}

	ws := RepoWorkSpace{RootDir: dir}
	err = ws.load(manifestURL)
	if err != nil {
		return nil, err
	}

	return &ws, nil
}
