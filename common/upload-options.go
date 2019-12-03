package common

// UploadOptions is options for upload related methods.
type UploadOptions struct {
	AutoTopic    bool
	CodeReviewID string
	Description  string
	DestBranch   string // Target branch for code review
	Draft        bool
	Issue        string
	LocalBranch  string // New
	MockGitPush  bool
	NoCertChecks bool
	NoEmails     bool
	OldOid       string // New
	People       [][]string
	Private      bool
	PushOptions  []string
	ProjectName  string // New
	ReviewURL    string // New
	Title        string
	UserEmail    string // New
	Version      int    // version: 1
	WIP          bool
}
