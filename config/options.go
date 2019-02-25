package config

// FetchOptions defines arguments for project.Fetch methods
type FetchOptions struct {
	Quiet             bool
	IsNew             bool
	CurrentBranchOnly bool
	ForceSync         bool
	CloneBundle       bool
	NoTags            bool
	Archive           bool
	OptimizedFetch    bool
	Prune             bool
	Submodules        bool
}
