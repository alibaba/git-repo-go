package project

import (
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"github.com/jiangxin/goconfig"
)

// RepoSettings holds settings in manifest project.
type RepoSettings struct {
	TopDir       string
	ManifestURL  string
	ManifestName string
	Groups       string
	Reference    string
	Revision     string
	Depth        int
	Archive      bool
	Dissociate   bool
	Mirror       bool
	Submodules   bool
	Config       goconfig.GitConfig
}

// ManifestProject is a special type of project.
type ManifestProject struct {
	Project
}

// ReadSettings reads settings from manifest project.
func (v *ManifestProject) ReadSettings() *RepoSettings {
	cfg := v.Config()

	s := v.Settings
	s.ManifestURL = cfg.Get(config.CfgRemoteOriginURL)
	s.ManifestName = cfg.Get(config.CfgManifestName)
	s.Groups = cfg.Get(config.CfgManifestGroups)
	s.Reference = cfg.Get(config.CfgRepoReference)
	s.Depth = cfg.GetInt(config.CfgRepoDepth, 0)
	s.Archive = cfg.GetBool(config.CfgRepoArchive, false)
	s.Dissociate = cfg.GetBool(config.CfgRepoDissociate, false)
	s.Mirror = cfg.GetBool(config.CfgRepoMirror, false)
	s.Submodules = cfg.GetBool(config.CfgRepoSubmodules, false)
	s.Config = v.Config()

	return s
}

// SaveSettings saves settings to manifest project.
func (v *ManifestProject) SaveSettings(s *RepoSettings) error {
	cfg := v.Config()

	v.Settings = s

	if s.ManifestURL != "" {
		cfg.Set(config.CfgRemoteOriginURL, s.ManifestURL)
	}

	if s.ManifestName != "" {
		cfg.Set(config.CfgManifestName, s.ManifestName)
	}

	if s.Groups != "" {
		cfg.Set(config.CfgManifestGroups, s.Groups)
	} else {
		cfg.Unset(config.CfgManifestGroups)
	}

	if s.Reference != "" {
		cfg.Set(config.CfgRepoReference, s.Reference)
	} else {
		cfg.Unset(config.CfgRepoReference)
	}

	if s.Depth > 0 {
		cfg.Set(config.CfgRepoDepth, s.Depth)
	} else {
		cfg.Unset(config.CfgRepoDepth)
	}

	// Only initialized for the first time, cannot unset
	if s.Archive {
		cfg.Set(config.CfgRepoArchive, true)
	}

	if s.Dissociate {
		cfg.Set(config.CfgRepoDissociate, true)
	} else {
		cfg.Unset(config.CfgRepoDissociate)
	}

	// Only initialized for the first time, cannot unset
	if s.Mirror {
		cfg.Set(config.CfgRepoMirror, true)
	}

	if s.Submodules {
		cfg.Set(config.CfgRepoSubmodules, true)
	} else {
		cfg.Unset(config.CfgRepoSubmodules)
	}

	return v.SaveConfig(cfg)
}

// MirrorEnabled checks if config variable repo.mirror is true.
func (v ManifestProject) MirrorEnabled() bool {
	return v.Config().GetBool(config.CfgRepoMirror, false)
}

// SubmoduleEnabled checks if config variable repo.submodules is true.
func (v ManifestProject) SubmoduleEnabled() bool {
	return v.Config().GetBool(config.CfgRepoSubmodules, false)
}

// ArchiveEnabled checks if config variable repo.archive is true.
func (v ManifestProject) ArchiveEnabled() bool {
	return v.Config().GetBool(config.CfgRepoArchive, false)
}

// DissociateEnabled checks if config variable repo.dissociate is true.
func (v ManifestProject) DissociateEnabled() bool {
	return v.Config().GetBool(config.CfgRepoDissociate, false)
}

// SetRevision changes project default branch.
func (v *ManifestProject) SetRevision(rev string) {
	v.Revision = rev
}

// NewManifestProject returns a manifest project: a worktree with a seperate repository.
func NewManifestProject(topDir, mURL string) *ManifestProject {
	p := ManifestProject{
		Project: *(NewProject(manifest.ManifestsProject,
			&RepoSettings{
				TopDir:      topDir,
				ManifestURL: mURL,
			})),
	}
	p.ReadSettings()
	if mURL != "" && p.Settings.ManifestURL == "" {
		p.Settings.ManifestURL = mURL
	}
	return &p
}
