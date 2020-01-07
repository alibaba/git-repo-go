package workspace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/errors"
	"github.com/aliyun/git-repo-go/project"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func testCreateManifests(topDir, manifestURL string) error {
	var (
		err error
	)

	// Create manifests.git repository
	mProject := project.NewManifestProject(topDir, manifestURL)
	err = mProject.GitInit()
	if err != nil {
		return fmt.Errorf("GitInit error: %s", err)
	}

	// Create manifests workdir
	err = mProject.PrepareWorkdir()
	if err != nil {
		return err
	}

	// Create manifest XMLs in manifest workdir
	filename := filepath.Join(mProject.WorkDir, "m1.xml")
	err = ioutil.WriteFile(filename, []byte(`<manifest>
  <remote name="aone" alias="origin" fetch=".." review="https://example.com" revision="default"></remote>
  <default remote="aone" revision="master"></default>
  <project name="platform/drivers" path="platform-drivers">
    <project name="platform/nic" path="nic"></project>
    <copyfile src="Makefile" dest="../Makefile"></copyfile>
  </project>
  <project name="platform/manifest" path="platform-manifest"></project>
</manifest>`), 0644)
	if err != nil {
		return err
	}

	filename = filepath.Join(mProject.WorkDir, "m2.xml")
	err = ioutil.WriteFile(filename, []byte(`<manifest>
  <remote name="origin" alias="origin" fetch=".." review="https://example.com" revision="default"></remote>
  <default remote="origin" revision="master"></default>
  <project name="jiangxin/hello" path="hello"/>
  <project name="jiangxin/world" path="world"/>
</manifest>`), 0644)
	if err != nil {
		return err
	}

	w, err := mProject.GitWorktree()
	if err != nil {
		return err
	}

	for _, f := range []string{"m1.xml", "m2.xml"} {
		_, err = w.Add(f)
		if err != nil {
			return err
		}
	}
	_, err = w.Commit("initial manifest", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Jiang Xin",
			Email: "worldhello.net@gmail.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func TestNewLoadEmptyRepoWorkSpace(t *testing.T) {
	var (
		tmpdir string
		err    error
		assert = assert.New(t)
	)

	tmpdir, err = ioutil.TempDir("", "git-repo-")
	if err != nil {
		panic(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workdir := filepath.Join(tmpdir, "workdir")
	err = os.MkdirAll(workdir, 0755)
	assert.Nil(err)

	ws, err := NewRepoWorkSpace(workdir)
	assert.Equal(errors.ErrRepoDirNotFound, err)
	assert.Nil(ws)
}

func TestNewLoadEmptyRepoWorkSpaceInit(t *testing.T) {
	var (
		tmpdir string
		err    error
		assert = assert.New(t)
	)

	tmpdir, err = ioutil.TempDir("", "git-repo-")
	if err != nil {
		panic(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workdir := filepath.Join(tmpdir, "workdir")
	err = os.MkdirAll(workdir, 0755)
	assert.Nil(err)

	ws, err := NewEmptyRepoWorkSpace(workdir, "git@server:path/of/manifest.git")
	assert.Nil(err)
	assert.Equal(workdir, ws.RootDir)
	assert.Nil(ws.Manifest)
	assert.NotNil(ws.ManifestProject)
	assert.Equal(0, len(ws.Projects))
}

func TestLoadRepoWorkSpace(t *testing.T) {
	var (
		tmpdir string
		err    error
		assert = assert.New(t)
	)

	tmpdir, err = ioutil.TempDir("", "git-repo-")
	if err != nil {
		panic(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workdir := filepath.Join(tmpdir, "workdir")
	err = os.MkdirAll(workdir, 0755)
	assert.Nil(err)
	mURL := "https://example.com/zhiyou.jx/manifest.git"
	err = testCreateManifests(workdir, mURL)
	assert.Nil(err)

	// No symlink: manifest.xml and no manifests/default.xml
	ws, err := NewRepoWorkSpace(workdir)
	assert.NotNil(err)
	assert.Nil(ws)

	// Create symlink
	src := filepath.Join(workdir, ".repo", "manifests", "m1.xml")
	target := filepath.Join(workdir, ".repo", "manifest.xml")
	err = os.Symlink(src, target)
	assert.Nil(err)
	ws, err = NewRepoWorkSpace(workdir)
	assert.Nil(err)
	assert.NotNil(ws.Manifest)
	assert.Equal(3, len(ws.Projects))

	// Create symlink
	assert.Nil(os.Remove(target))
	src = filepath.Join(workdir, ".repo", "manifests", "m2.xml")
	err = os.Symlink(src, target)
	assert.Nil(err)
	ws.load("")
	assert.Equal(2, len(ws.Projects))

	// Create manifest settings
	assert.Nil(os.Remove(target))
	cfg := ws.ManifestProject.Config()
	cfg.Set(config.CfgManifestName, "m1.xml")
	assert.Nil(ws.ManifestProject.SaveConfig(cfg))
	ws.load("")
	assert.Equal(3, len(ws.Projects))
}

func TestManifestsProjectName(t *testing.T) {
	var (
		expect string
		actual string
		URL    = "https://example.com/my/test/repository"
	)

	assert := assert.New(t)

	expect = "manifests"
	actual = manifestsProjectName("", ".")
	assert.Equal(expect, actual)

	expect = "manifests"
	actual = manifestsProjectName(URL, "")
	assert.Equal(expect, actual)

	expect = "repository"
	actual = manifestsProjectName(URL, ".")
	assert.Equal(expect, actual)

	expect = "test/repository"
	actual = manifestsProjectName(URL, "..")
	assert.Equal(expect, actual)

	expect = "my/test/repository"
	actual = manifestsProjectName(URL, "../..")
	assert.Equal(expect, actual)

	expect = "my/test/repository"
	actual = manifestsProjectName(URL, "../../")
	assert.Equal(expect, actual)

	expect = "example.com/my/test/repository"
	actual = manifestsProjectName(URL, "../../../../../../..")
	assert.Equal(expect, actual)
}
