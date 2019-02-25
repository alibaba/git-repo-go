package workspace

import (
	"os"
	"path/filepath"

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
	ManifestProject *project.Project
	Projects        []*project.Project
	reference       string
}

// IsInitialized checks whether workspace is initialized
func (v *WorkSpace) IsInitialized() bool {
	if v.RootDir == "" || v.Manifest == nil || v.ManifestProject == nil {
		return false
	}

	if v.ManifestProject.RemoteURL() == "" {
		return false
	}

	return true
}

// SetReference set reference for workspace
func (v *WorkSpace) SetReference(reference string) {
	v.reference = reference
}

// ManifestURL returns URL of manifest project
func (v *WorkSpace) ManifestURL() string {
	if v.ManifestProject == nil {
		return ""
	}
	return v.ManifestProject.RemoteURL()
}

// GetReference returns reference
func (v *WorkSpace) GetReference() string {
	return v.reference
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
func (v *WorkSpace) LinkManifest(name string) error {
	if name != "" {
		cfg := v.Config()
		if cfg.Get(config.CfgManifestName) != name {
			cfg.Set(config.CfgManifestName, name)
			v.SaveConfig(cfg)
		}

		if cap.Symlink() {
			target := filepath.Join(v.RootDir, config.DotRepo, config.ManifestXML)
			src, err := os.Readlink(target)
			if err != nil || filepath.Base(src) != name {
				os.Remove(target)
				src = filepath.Join(config.Manifests, name)
				log.Errorf("will symlink '%s' to '%s'", src, target)
				err = os.Symlink(src, target)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Load will read manifest file and repository for workspace
func (v *WorkSpace) Load(manifestURL string) error {
	m, err := manifest.Load(filepath.Join(v.RootDir, config.DotRepo))
	if err != nil {
		return err
	}

	v.Manifest = m

	// Set manifest project even v.Manifest is nil
	v.ManifestProject = project.NewManifestProject(v.RootDir, manifestURL)
	manifestURL = v.ManifestProject.RemoteURL()

	// Set projects
	v.Projects = []*project.Project{}
	if v.Manifest != nil {
		for _, p := range v.Manifest.AllProjects() {
			v.Projects = append(v.Projects,
				project.NewProject(&p, v.RootDir, manifestURL))
		}
	}

	return nil
}

// NewWorkSpace finds and loads repo root
func NewWorkSpace(dir string) (*WorkSpace, error) {
	return NewWorkSpace2(dir, "")
}

// NewWorkSpace2 finds repo root and load specific manifest file
func NewWorkSpace2(dir, manifestURL string) (*WorkSpace, error) {
	var (
		err error
	)

	repoRoot, err := path.FindRepoRoot(dir)
	if err != nil && err != errors.ErrRepoDirNotFound {
		return nil, err
	}
	if repoRoot == "" {
		dir, err = path.Abs(dir)
		if err != nil {
			return nil, err
		}
		repoRoot = dir
	}

	ws := WorkSpace{RootDir: repoRoot}
	err = ws.Load(manifestURL)
	if err != nil {
		return nil, err
	}

	return &ws, nil
}
