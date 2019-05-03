package workspace

import (
	"gopkg.in/h2non/gock.v1"
	"testing"

	"code.alibaba-inc.com/force/git-repo/manifest"
	"github.com/stretchr/testify/assert"
)

func TestLoadRemoteReviewUrlHasGerritSuffix(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com/Gerrit",
		Revision: "master",
		Type:     "",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString("ssh.example.com 22")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote.GetSSHInfo())
	assert.Equal("ssh.example.com", remote.GetSSHInfo().Host)
	assert.Equal(22, remote.GetSSHInfo().Port)
	assert.Equal("gerrit", remote.GetType())
}

func TestLoadRemoteReviewUrlHasAGitSuffix(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com/AGit",
		Revision: "master",
		Type:     "",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString("ssh.example.com 22")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote.GetSSHInfo())
	assert.Equal("ssh.example.com", remote.GetSSHInfo().Host)
	assert.Equal(22, remote.GetSSHInfo().Port)
	assert.Equal("agit", remote.GetType())
}

func TestLoadRemoteSSHInfoDefaultGerrit(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com",
		Revision: "master",
		Type:     "",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString("ssh.example.com 22")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote.GetSSHInfo())
	assert.Equal("ssh.example.com", remote.GetSSHInfo().Host)
	assert.Equal(22, remote.GetSSHInfo().Port)
	assert.Equal("gerrit", remote.GetType())
}

func TestLoadRemoteManifestOverrideDefaultType(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com",
		Revision: "master",
		Type:     "agit",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString("ssh.example.com 22")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote.GetSSHInfo())
	assert.Equal("ssh.example.com", remote.GetSSHInfo().Host)
	assert.Equal(22, remote.GetSSHInfo().Port)
	assert.Equal("agit", remote.GetType())
}

func TestLoadRemoteBadSSHInfo(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com",
		Revision: "master",
		Type:     "agit",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString("ssh.example.com")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.NotNil(err)
	assert.Nil(remote)
}

func TestLoadRemoteJSON(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com",
		Revision: "master",
		Type:     "",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString(`{ "host": "ssh.example.com", "port": 22, "type": "agit" }`)

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote)
	assert.Equal("ssh.example.com", remote.GetSSHInfo().Host)
	assert.Equal(22, remote.GetSSHInfo().Port)
	assert.Equal("agit", remote.GetType())
}

func TestLoadRemoteEmptyReview(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "",
		Revision: "master",
		Type:     "",
	}

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote)
	assert.Nil(remote.GetSSHInfo())
	assert.Equal("unknown", remote.GetType())
	assert.Equal("origin", remote.GetRemote().Name)
	assert.Equal("", remote.GetRemote().Review)
}

func TestLoadRemoteSSHProtocolReview(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "ssh://git@example.com",
		Revision: "master",
		Type:     "",
	}

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote)
	assert.Nil(remote.GetSSHInfo())
	assert.Equal("gerrit", remote.GetType())
	assert.Equal("ssh://git@example.com", remote.GetRemote().Review)
}

func TestLoadRemoteNotAvailable(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com",
		Revision: "master",
		Type:     "",
	}

	gock.New("https://example.com").
		Reply(200).
		BodyString("NOT_AVAILABLE")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote)
	assert.Nil(remote.GetSSHInfo())
	assert.Equal("gerrit", remote.GetType())
	assert.Equal("https://example.com", remote.GetRemote().Review)
}

func TestLoadRemoteHTTPStatus404(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
	)

	mRemote := manifest.Remote{
		Name:     "origin",
		Fetch:    "..",
		PushURL:  "",
		Review:   "https://example.com",
		Revision: "master",
		Type:     "",
	}

	gock.New("https://example.com").
		Reply(404).
		BodyString("Not Found")

	client := getHTTPClient()
	gock.InterceptClient(client)

	remote, err := loadRemote(&mRemote)
	assert.Nil(err)
	assert.NotNil(remote)
	assert.Nil(remote.GetSSHInfo())
	assert.Equal("unknown", remote.GetType())
	assert.Equal("https://example.com", remote.GetRemote().Review)
}
