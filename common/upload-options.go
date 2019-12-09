package common

// UploadOptions is options for upload related methods.
type UploadOptions struct {
	AutoTopic    bool
	CodeReview   CodeReview
	Description  string
	DestBranch   string // Target branch for code review
	Draft        bool
	Issue        string
	LocalBranch  string // New
	MockGitPush  bool
	NoCertChecks bool
	NoEmails     bool
	OldOid       string
	People       [][]string
	Private      bool
	PushOptions  []string
	ProjectName  string // New
	RemoteName   string
	RemoteURL    string
	Title        string
	Version      int // version: 1
	WIP          bool
}
