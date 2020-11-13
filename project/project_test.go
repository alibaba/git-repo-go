package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/alibaba/git-repo-go/manifest"
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
		RemoteName: "origin",
		Revision:   "refs/heads/master",
		DestBranch: "refs/heads/master",
	}
	xmlProject.ManifestRemote = &manifest.Remote{
		Name:  "origin",
		Fetch: "..",
	}
	p := NewProject(&xmlProject,
		&RepoSettings{
			TopDir:      workdir,
			ManifestURL: mURL,
		}, nil)
	u, err := p.GetRemoteURL()
	assert.Nil(err)
	assert.Equal("https://github.com/my/foo.git", u)

	// Call GitInit
	assert.False(p.IsRepoInitialized())
	err = p.GitInit()
	assert.Nil(err)
	// TODO: fix it
	assert.Equal("https://github.com/my/foo.git",
		p.gitConfigRemoteURL())
	return

	u, err = p.GetRemoteURL()
	assert.Nil(err)
	assert.Equal("https://github.com/jiangxin/my/foo.git", u)

	// Call GitInit twice
	mURL = "https://example.com/zhiyou.jx/manifest.git"
	p.SetManifestURL(mURL)
	err = p.GitInit()
	assert.Nil(err)
	// TODO: fix it
	assert.Equal("https://example.com/zhiyou.jx/my/foo.git",
		p.gitConfigRemoteURL())
	u, err = p.GetRemoteURL()
	assert.Equal("https://example.com/zhiyou.jx/my/foo.git", u)
}

func TestProjectMatchGroups(t *testing.T) {
	var mGroups string
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
		RemoteName: "origin",
		Revision:   "refs/heads/master",
		DestBranch: "refs/heads/master",
	}
	xmlProject.ManifestRemote = &manifest.Remote{
		Name:  "origin",
		Fetch: "..",
	}
	p := NewProject(&xmlProject,
		&RepoSettings{
			TopDir:      workdir,
			ManifestURL: mURL,
		}, nil)

	p.Groups = ""
	assert.True(p.MatchGroups(""))
	assert.True(p.MatchGroups("default"))
	assert.True(p.MatchGroups("all"))

	p.Groups = "group1,group2"
	mGroups = "group3"
	assert.False(p.MatchGroups(mGroups))

	p.Groups = "group1,group2"
	mGroups = "-group1,group2"
	assert.True(p.MatchGroups(mGroups))

	p.Groups = "group1,group2"
	mGroups = "group1,-group2"
	assert.False(p.MatchGroups(mGroups))

	p.Groups = "notdefault"
	mGroups = ""
	assert.False(p.MatchGroups(mGroups))

	p.Groups = "notdefault,group1"
	mGroups = ""
	assert.False(p.MatchGroups(mGroups))

	p.Groups = "notdefault,group1"
	mGroups = "default"
	assert.False(p.MatchGroups(mGroups))

	p.Groups = "notdefault,group1"
	mGroups = "group1"
	assert.True(p.MatchGroups(mGroups))

	p.Groups = "notdefault,group1"
	mGroups = "group2"
	assert.False(p.MatchGroups(mGroups))

	p.Groups = "notdefault,group1"
	mGroups = "all"
	assert.True(p.MatchGroups(mGroups))
}

func TestIndexByName(t *testing.T) {
	assert := assert.New(t)
	projects := []*Project{
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group3/Name1",
					Path: "app/3-1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group1/Name1",
					Path: "app/1-1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/Name2",
					Path: "app/2-2-1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/Name1",
					Path: "app/2-1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/Name2",
					Path: "app/2-2-2",
				},
			},
		},
	}
	expect := `name: Group1/Name1, path-1: app/1-1
name: Group2/Name1, path-1: app/2-1
name: Group2/Name2, path-1: app/2-2-1
name: Group2/Name2, path-2: app/2-2-2
name: Group3/Name1, path-1: app/3-1`
	actual := []string{}
	for name, ps := range IndexByName(projects) {
		for i, p := range ps {
			actual = append(actual, fmt.Sprintf("name: %s, path-%d: %s", name, i+1, p.Path))
		}
	}
	sort.Strings(actual)
	assert.Equal(expect, strings.Join(actual, "\n"))
}

func TestProjectsTree(t *testing.T) {
	assert := assert.New(t)
	projects := []*Project{
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group3/Name1",
					Path: "app3/name1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group1/Name1",
					Path: "app1/name1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/App",
					Path: "app2",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/Name2/sub1",
					Path: "app2/name2/sub1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/Name1",
					Path: "app2/name1",
				},
			},
		},
		&Project{
			Repository: Repository{
				Project: manifest.Project{
					Name: "Group2/Name2",
					Path: "app2/name2",
				},
			},
		},
	}
	expect := `path: app1/name1, name: Group1/Name1
path: app2, name: Group2/App
path: app3/name1, name: Group3/Name1`
	actual := []string{}
	tree := ProjectsTree(projects)
	for _, e := range tree.Trees {
		actual = append(actual, fmt.Sprintf("path: %s, name: %s", e.Project.Path, e.Project.Name))
	}
	assert.Equal(expect, strings.Join(actual, "\n"))

	expect = `path: app2/name1, name: Group2/Name1
path: app2/name2, name: Group2/Name2`
	actual = []string{}
	for _, e := range tree.Trees[1].Trees {
		actual = append(actual, fmt.Sprintf("path: %s, name: %s", e.Project.Path, e.Project.Name))
	}
	assert.Equal(expect, strings.Join(actual, "\n"))

	expect = "path: app2/name2/sub1, name: Group2/Name2/sub1"
	actual = []string{}
	for _, e := range tree.Trees[1].Trees[1].Trees {
		actual = append(actual, fmt.Sprintf("path: %s, name: %s", e.Project.Path, e.Project.Name))
	}
	assert.Equal(expect, strings.Join(actual, "\n"))
}
