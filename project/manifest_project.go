package project

import (
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

// ManifestProject is a special type of project
type ManifestProject struct {
	Project
}

// MirrorEnabled checks if config variable repo.mirror is true
func (v ManifestProject) MirrorEnabled() bool {
	b, _ := v.Config().GetBool(config.CfgRepoMirror, false)
	return b
}

// SubmoduleEnabled checks if config variable repo.submodules is true
func (v ManifestProject) SubmoduleEnabled() bool {
	b, _ := v.Config().GetBool(config.CfgRepoSubmodules, false)
	return b
}

// ArchiveEnabled checks if config variable repo.archive is true
func (v ManifestProject) ArchiveEnabled() bool {
	b, _ := v.Config().GetBool(config.CfgRepoArchive, false)
	return b
}

// DissociateEnabled checks if config variable repo.dissociate is true
func (v ManifestProject) DissociateEnabled() bool {
	b, _ := v.Config().GetBool(config.CfgRepoDissociate, false)
	return b
}

// NewManifestProject returns a manifest project: a worktree with a seperate repository
func NewManifestProject(repoRoot, manifestURL string) *ManifestProject {
	p := NewProject(manifest.ManifestsProject, repoRoot, manifestURL)
	return &ManifestProject{Project: *p}
}
