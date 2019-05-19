package manifest

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jiangxin/multi-log"
	"github.com/stretchr/testify/assert"
)

var (
	manifest Manifest
	xmlData  []byte
)

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	manifest = Manifest{
		Remotes: []Remote{
			Remote{
				Name:     "aone",
				Alias:    "origin",
				Fetch:    "https://code.aone.alibaba-inc.com",
				Review:   "https://code.aone.alibaba-inc.com",
				Revision: "default",
			},
		},
		Default: &Default{
			RemoteName: "origin",
			Revision:   "master",
		},
		Projects: []Project{
			Project{
				Name: "platform/drivers",
				Path: "platform-drivers",
				Projects: []Project{
					Project{
						Name: "platform/nic",
						Path: "nic",
					},
				},
				CopyFiles: []CopyFile{
					CopyFile{
						Src:  "Makefile",
						Dest: "../Makefile",
					},
				},
			},
			Project{
				Name: "platform/manifest",
				Path: "platform-manifest",
			},
		},
	}

	expected := `<manifest>
  <remote name="aone" alias="origin" fetch="https://code.aone.alibaba-inc.com" review="https://code.aone.alibaba-inc.com" revision="default"></remote>
  <default remote="origin" revision="master"></default>
  <project name="platform/drivers" path="platform-drivers">
    <project name="platform/nic" path="nic"></project>
    <copyfile src="Makefile" dest="../Makefile"></copyfile>
  </project>
  <project name="platform/manifest" path="platform-manifest"></project>
</manifest>`

	actual, err := xml.MarshalIndent(&manifest, "", "  ")
	xmlData = actual
	assert.Nil(err)
	assert.Equal(expected, string(actual))
}

func TestUnmarshal(t *testing.T) {
	assert := assert.New(t)

	m := Manifest{}
	err := xml.Unmarshal(xmlData, &m)

	assert.Nil(err)
	assert.Equal(manifest.Default, m.Default)
	assert.Equal(manifest.Remotes, m.Remotes)
	assert.Equal(manifest.Projects, m.Projects)
}

func TestLoad(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo")
	if err != nil {
		log.Fatal(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workDir := filepath.Join(tmpdir, "workdir")
	repoDir := filepath.Join(workDir, ".repo")
	manifestDir := filepath.Join(repoDir, ".repo", "manifests")
	err = os.MkdirAll(manifestDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// load with missing manifest will fail
	m, err := Load(repoDir)
	assert.NotNil(err)
	assert.Nil(m)

	// create manifest.xml
	manifestFile := filepath.Join(repoDir, "manifest.xml")
	err = ioutil.WriteFile(manifestFile, []byte(`
<manifest>
  <remote name="aone" alias="origin" fetch="https://code.aone.alibaba-inc.com" review="https://code.aone.alibaba-inc.com" revision="default"></remote>
  <default remote="aone" revision="master"></default>
  <project name="platform/drivers" path="platform-drivers">
    <project name="nic" path="nic"></project>
    <copyfile src="Makefile" dest="../Makefile"></copyfile>
  </project>
  <project name="platform/manifest" path="platform-manifest"></project>
</manifest>`), 0644)
	assert.Nil(err)

	m, err = Load(repoDir)
	assert.Nil(err)
	assert.NotNil(m)
	assert.Equal(
		&Default{
			RemoteName: "aone",
			Revision:   "master",
		}, m.Default)
	assert.Equal(
		[]Remote{Remote{
			Name:     "aone",
			Alias:    "origin",
			Fetch:    "https://code.aone.alibaba-inc.com",
			Review:   "https://code.aone.alibaba-inc.com",
			Revision: "default"},
		}, m.Remotes)
	projects := []string{}
	for _, p := range m.Projects {
		projects = append(projects, p.Name)
	}
	assert.Equal([]string{
		"platform/drivers",
		"platform/drivers/nic",
		"platform/manifest"},
		projects)
}

func TestInclude(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo")
	if err != nil {
		log.Fatal(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workDir := filepath.Join(tmpdir, "workdir")
	repoDir := filepath.Join(workDir, ".repo")
	manifestDir := filepath.Join(repoDir, ".repo", "manifests")
	err = os.MkdirAll(manifestDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// create manifest.xml
	manifestFile := filepath.Join(repoDir, "manifest.xml")
	err = ioutil.WriteFile(manifestFile, []byte(`
<manifest>
  <remote name="aone" alias="origin" fetch="https://code.aone.alibaba-inc.com" review="https://code.aone.alibaba-inc.com" revision="default"></remote>
  <default remote="aone" revision="master"></default>
  <project name="platform/drivers" path="platform-drivers">
    <project name="nic" path="nic"></project>
    <copyfile src="Makefile" dest="../Makefile"></copyfile>
  </project>
  <project name="platform/manifest" path="platform-manifest"></project>
  <include name="../manifest.inc"></include>
</manifest>`), 0644)
	assert.Equal(nil, err)

	err = ioutil.WriteFile(filepath.Join(workDir, "manifest.inc"), []byte(`
<manifest>
  <project name="platform/foo" path="foo"/>
</manifest>`), 0644)
	assert.Equal(nil, err)

	m, err := Load(repoDir)
	assert.Equal(nil, err)
	projects := []string{}
	for _, p := range m.Projects {
		projects = append(projects, p.Name)
	}
	assert.Equal([]string{
		"platform/drivers",
		"platform/drivers/nic",
		"platform/manifest",
		"platform/foo"},
		projects)

	// all project has valid remote
	for _, p := range m.AllProjects() {
		assert.NotNil(p.ManifestRemote)
	}
}

func TestLoadWithLocalManifest(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo")
	if err != nil {
		log.Fatal(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workDir := filepath.Join(tmpdir, "workdir")
	repoDir := filepath.Join(workDir, ".repo")
	manifestDir := filepath.Join(repoDir, ".repo", "manifests")
	err = os.MkdirAll(manifestDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// create manifest.xml
	manifestFile := filepath.Join(repoDir, "manifest.xml")
	err = ioutil.WriteFile(manifestFile, []byte(`
<manifest>
  <remote name="aone" alias="origin" fetch="https://code.aone.alibaba-inc.com" review="https://code.aone.alibaba-inc.com" revision="default"></remote>
  <default remote="aone" revision="master"></default>
  <project name="platform/drivers" path="platform-drivers">
    <project name="nic" path="nic"></project>
    <copyfile src="Makefile" dest="../Makefile"></copyfile>
  </project>
  <project name="platform/manifest" path="platform-manifest"></project>
</manifest>`), 0644)
	assert.Equal(nil, err)

	// create local_manifest.xml
	localManifestFile := filepath.Join(repoDir, "local_manifest.xml")
	err = ioutil.WriteFile(localManifestFile, []byte(`
<manifest>
  <remove-project name="platform/manifest"></remove-project>
</manifest>`), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// create local_manifests/test.xml
	localManifestFile = filepath.Join(repoDir, "local_manifests", "test.xml")
	os.MkdirAll(filepath.Dir(localManifestFile), 0755)
	err = ioutil.WriteFile(localManifestFile, []byte(`
<manifest>
  <project name="tools/git-repo" path="tools/git-repo"/>
</manifest>`), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// load all manifest and test
	m, err := Load(repoDir)
	assert.Equal(nil, err)
	projects := []string{}
	for _, p := range m.Projects {
		projects = append(projects, p.Name)
	}
	assert.Equal([]string{
		"platform/drivers",
		"platform/drivers/nic",
		"tools/git-repo"}, projects)
}

func TestCircularInclude(t *testing.T) {
	assert := assert.New(t)

	tmpdir, err := ioutil.TempDir("", "git-repo")
	if err != nil {
		log.Fatal(err)
	}
	defer func(dir string) {
		os.RemoveAll(dir)
	}(tmpdir)

	workDir := filepath.Join(tmpdir, "workdir")
	repoDir := filepath.Join(workDir, ".repo")
	manifestDir := filepath.Join(repoDir, ".repo", "manifests")
	err = os.MkdirAll(manifestDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// create manifest.xml
	manifestFile := filepath.Join(repoDir, "manifest.xml")
	err = ioutil.WriteFile(manifestFile, []byte(`
<manifest>
  <remote name="aone" alias="origin" fetch="https://code.aone.alibaba-inc.com" review="https://code.aone.alibaba-inc.com" revision="default"></remote>
  <default remote="aone" revision="master"></default>
  <project name="platform/drivers" path="platform-drivers">
    <project name="platform/nic" path="nic"></project>
    <copyfile src="Makefile" dest="../Makefile"></copyfile>
  </project>
  <project name="platform/manifest" path="platform-manifest"></project>
  <include name="../manifest.inc"></include>
</manifest>`), 0644)
	assert.Equal(nil, err)

	err = ioutil.WriteFile(filepath.Join(workDir, "manifest.inc"), []byte(`
<manifest>
  <project name="platform/foo" path="foo"/>
  <include name="manifest2.inc"/>
</manifest>`), 0644)
	assert.Equal(nil, err)

	err = ioutil.WriteFile(filepath.Join(workDir, "manifest2.inc"), []byte(`
<manifest>
  <project name="platform/bar" path="bar"/>
  <include name="manifest.inc"/>
</manifest>`), 0644)
	assert.Equal(nil, err)

	m, err := Load(repoDir)
	assert.Contains(err.Error(), "exceeded maximum include depth (10)")
	assert.Equal(true, nil == m)
}
