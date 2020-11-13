package config

// UploadOptions is options for upload related methods.
type UploadOptions struct {
	AutoTopic    bool
	CodeReview   CodeReview // Directly edit remote code review.
	Description  string
	DestBranch   string // Target branch for code review.
	Draft        bool
	Issue        string
	LocalBranch  string // Local branch with commits, will push to remote.
	MockGitPush  bool
	NoCertChecks bool
	NoEmails     bool
	OldOid       string
	People       [][]string
	Private      bool
	PushOptions  []string
	RemoteName   string
	RemoteURL    string
	Title        string
	WIP          bool
}
