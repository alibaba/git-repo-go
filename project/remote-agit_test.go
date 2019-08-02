package project

import (
	"testing"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/manifest"
	"github.com/stretchr/testify/assert"
)

type mockGitWithPushOptions struct {
}

func (v mockGitWithPushOptions) GitCanPushOptions() bool {
	return true
}

type mockGitWithoutPushOptions struct {
}

func (v mockGitWithoutPushOptions) GitCanPushOptions() bool {
	return false
}

func TestUploadCommandsWithPushOptions(t *testing.T) {
	var (
		mockGitCap = mockGitWithPushOptions{}
		assert     = assert.New(t)
		err        error
	)

	cap.CapGit = mockGitCap

	mURL := "https://code.alibaba-inc.com/my/foo.git"
	xmlProject := manifest.Project{
		Name:       "my/foo",
		Path:       "dir/foo",
		RemoteName: "origin",
		Revision:   "refs/heads/master",
		DestBranch: "refs/heads/master",
	}

	xmlRemote := manifest.Remote{
		Name:  "origin",
		Fetch: "..",
	}

	xmlProject.ManifestRemote = &xmlRemote

	p := NewProject(&xmlProject, &RepoSettings{
		RepoRoot:    "/dev/null",
		ManifestURL: mURL,
	})

	branch := ReviewableBranch{
		Project: p,
		Branch: Branch{
			Name: "refs/heads/my/topic",
			Hash: "1234",
		},
		DestBranch: "master",
		RemoteTrack: Reference{
			Name: "refs/remotes/origin/master",
			Hash: "2345",
		},
		Uploaded: false,
	}

	remote := AGitRemote{
		Remote: xmlRemote,
		SSHInfo: &SSHInfo{
			Host: "code.alibaba-inc.com",
			Port: 22,
			Type: "agit",
		},
	}

	options := UploadOptions{
		Description:  "description of the MR...",
		DestBranch:   "master",
		Draft:        false,
		Issue:        "123",
		MockGitPush:  true,
		NoCertChecks: true,
		NoEmails:     false,
		People: [][]string{
			[]string{
				"user1",
				"user2",
			},
			[]string{
				"user3",
			},
		},
		Private:     false,
		PushOptions: []string{},
		Title:       "title of the MR",
		UserEmail:   "author@foo.bar",
		WIP:         false,
	}
	cmds, err := remote.UploadCommands(&options, &branch)
	assert.Nil(err)

	expect := []string{
		"git",
		"push",
		"--receive-pack=agit-receive-pack",
		"-o",
		"title=title of the MR",
		"-o",
		"description=description of the MR...",
		"-o",
		"issue=123",
		"-o",
		"reviewers=user1,user2",
		"-o",
		"cc=user3",
		"ssh://git@code.alibaba-inc.com/my/foo.git",
		"refs/heads/my/topic:refs/for/master/my/topic",
	}

	assert.Equal(expect, cmds)
}

func TestUploadCommandsWithoutPushOptions(t *testing.T) {
	var (
		mockGitCap = mockGitWithoutPushOptions{}
		assert     = assert.New(t)
		err        error
	)

	cap.CapGit = mockGitCap

	mURL := "https://code.alibaba-inc.com/my/foo.git"
	xmlProject := manifest.Project{
		Name:       "my/foo",
		Path:       "dir/foo",
		RemoteName: "origin",
		Revision:   "refs/heads/master",
		DestBranch: "refs/heads/master",
	}

	xmlRemote := manifest.Remote{
		Name:  "origin",
		Fetch: "..",
	}

	xmlProject.ManifestRemote = &xmlRemote

	p := NewProject(&xmlProject, &RepoSettings{
		RepoRoot:    "/dev/null",
		ManifestURL: mURL,
	})

	branch := ReviewableBranch{
		Project: p,
		Branch: Branch{
			Name: "refs/heads/my/topic",
			Hash: "1234",
		},
		DestBranch: "master",
		RemoteTrack: Reference{
			Name: "refs/remotes/origin/master",
			Hash: "2345",
		},
		Uploaded: false,
	}

	remote := AGitRemote{
		Remote: xmlRemote,
		SSHInfo: &SSHInfo{
			Host: "code.alibaba-inc.com",
			Port: 22,
			Type: "agit",
		},
	}

	options := UploadOptions{
		Description:  "description of the MR...",
		DestBranch:   "master",
		Draft:        true,
		Issue:        "123",
		MockGitPush:  true,
		NoCertChecks: true,
		NoEmails:     false,
		People: [][]string{
			[]string{
				"user1",
				"user2",
			},
			[]string{
				"user3",
			},
		},
		Private:     true,
		PushOptions: []string{},
		Title:       "title of the MR",
		UserEmail:   "author@foo.bar",
		WIP:         true,
	}
	cmds, err := remote.UploadCommands(&options, &branch)
	assert.Nil(err)

	expect := []string{
		"git",
		"push",
		"--receive-pack=agit-receive-pack",
		"ssh://git@code.alibaba-inc.com/my/foo.git",
		"refs/heads/my/topic:refs/drafts/master/my/topic%r=user1,r=user2,cc=user3,private,wip",
	}

	assert.Equal(expect, cmds)
}
