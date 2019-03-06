package path

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo")
	if err != nil {
		panic(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	home := os.Getenv("HOME")
	os.Setenv("HOME", tmpdir)

	assert.Equal(filepath.Join(tmpdir, ".config", "git"), xdgConfigHome(""))
	assert.Equal(filepath.Join(tmpdir, ".config", "git", "config"), xdgConfigHome("config"))

	assert.Equal(tmpdir, homeDir())

	assert.Equal(tmpdir, expendHome(""))
	assert.Equal(tmpdir, expendHome("~"))
	assert.Equal(tmpdir, expendHome("~/"))
	assert.Equal(filepath.Join(tmpdir, ".bashrc"), expendHome("~/.bashrc"))
	assert.Equal(filepath.Join(tmpdir, ".bashrc"), expendHome(".bashrc"))
	assert.Equal("/etc/profile.d", expendHome("/etc/profile.d"))

	cwd, _ := os.Getwd()
	assert.Equal(cwd, Abs(""))
	assert.Equal(filepath.Join(cwd, "dir/file"), Abs("dir/file"))
	assert.Equal("/etc/profile.d", Abs("/etc/profile.d"))
	assert.Equal(tmpdir, Abs("~"))
	assert.Equal(tmpdir, Abs("~/"))
	assert.Equal(filepath.Join(tmpdir, ".bashrc"), Abs("~/.bashrc"))

	cwd = "/some/dir"
	assert.Equal(cwd, AbsJoin(cwd, ""))
	assert.Equal(filepath.Join(cwd, "dir/file"), AbsJoin(cwd, "dir/file"))
	assert.Equal("/etc/profile.d", AbsJoin(cwd, "/etc/profile.d"))
	assert.Equal(tmpdir, AbsJoin(cwd, "~"))
	assert.Equal(tmpdir, AbsJoin(cwd, "~/"))
	assert.Equal(filepath.Join(tmpdir, ".bashrc"), AbsJoin(cwd, "~/.bashrc"))

	os.Setenv("HOME", home)
}

func TestFindRepoRoot(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo")
	if err != nil {
		panic(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	assert.Nil(os.MkdirAll(filepath.Join(tmpdir, "a", "b"), 0755))
	os.Chdir(filepath.Join(tmpdir, "a", "b"))

	assert.Equal("", FindRepoRoot(""))
	assert.Equal("", FindRepoRoot(tmpdir))

	assert.Nil(os.MkdirAll(filepath.Join(tmpdir, ".repo"), 0755))
	assert.Equal(tmpdir, FindRepoRoot(""))
	assert.Equal(tmpdir, FindRepoRoot(tmpdir))
	assert.Equal(tmpdir, FindRepoRoot(filepath.Join(tmpdir, "a", "b")))
}
