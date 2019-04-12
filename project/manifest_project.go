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
	return v.Config().GetBool(config.CfgRepoMirror, false)
}

// SubmoduleEnabled checks if config variable repo.submodules is true
func (v ManifestProject) SubmoduleEnabled() bool {
	return v.Config().GetBool(config.CfgRepoSubmodules, false)
}

// ArchiveEnabled checks if config variable repo.archive is true
func (v ManifestProject) ArchiveEnabled() bool {
	return v.Config().GetBool(config.CfgRepoArchive, false)
}

// DissociateEnabled checks if config variable repo.dissociate is true
func (v ManifestProject) DissociateEnabled() bool {
	return v.Config().GetBool(config.CfgRepoDissociate, false)
}

// NewManifestProject returns a manifest project: a worktree with a seperate repository
func NewManifestProject(repoRoot, manifestURL string) *ManifestProject {
	p := NewProject(manifest.ManifestsProject, repoRoot, manifestURL)
	return &ManifestProject{Project: *p}
}
