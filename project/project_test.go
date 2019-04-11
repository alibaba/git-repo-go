package project

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"code.alibaba-inc.com/force/git-repo/manifest"
	"github.com/stretchr/testify/assert"
)

func TestProjectGitInit(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo-")
	if err != nil {
		panic(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workdir := filepath.Join(tmpdir, "work")
	assert.Nil(os.MkdirAll(workdir, 0755))
	mURL := "https://github.com/jiangxin/manifest"

	xmlProject := manifest.Project{
		Name:       "my/foo",
		Path:       "dir/foo",
		Remote:     "origin",
		Revision:   "refs/heads/master",
		DestBranch: "refs/heads/master",
	}
	xmlProject.SetRemote(&manifest.Remote{
		Name:  "origin",
		Fetch: "..",
	})
	p := NewProject(&xmlProject, workdir, mURL)
	assert.Equal("https://github.com/jiangxin/my/foo", p.RemoteURL())

	// Call GitInit
	assert.False(p.IsRepoInitialized())
	err = p.GitInit(mURL, "")
	assert.Nil(err)
	assert.Equal("https://github.com/jiangxin/my/foo",
		p.GetGitRemoteURL())
	assert.Equal("https://github.com/jiangxin/my/foo",
		p.RemoteURL())

	// Call GitInit twice
	mURL = "https://code.aone.alibaba-inc.com/zhiyou.jx/manifest.git"
	err = p.GitInit(mURL, "")
	assert.Nil(err)
	assert.Equal("https://code.aone.alibaba-inc.com/zhiyou.jx/my/foo",
		p.GetGitRemoteURL())
	assert.Equal("https://code.aone.alibaba-inc.com/zhiyou.jx/my/foo",
		p.RemoteURL())
}
