package manifest

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	manifest Manifest
	xmlData  []byte
)

func TestUnmarshal1(t *testing.T) {
	var m Manifest
	assert := assert.New(t)
	data := `<?xml version="1.0" encoding="UTF-8"?>
<manifest>
  <project path="manifest"
           name="tools/manifest" />
  <project path="platform-manifest"
           name="platform/manifest" />
</manifest>`

	err := xml.Unmarshal([]byte(data), &m)
	assert.Nil(err)
	assert.Equal(2, len(m.Projects))
	assert.Equal("manifest", m.Projects[0].Path)
	assert.Equal("tools/manifest", m.Projects[0].Name)
	assert.Equal("platform-manifest", m.Projects[1].Path)
	assert.Equal("platform/manifest", m.Projects[1].Name)
}

func TestMarshal1(t *testing.T) {
	var m Manifest
	assert := assert.New(t)

	m = Manifest{
		Projects: []Project{
			Project{
				Name: "tools/manifest",
				Path: "manifest",
			},
			Project{
				Name: "platform/manifest",
				Path: "platform-manifest",
			},
		},
	}

	expected := `<manifest>
  <project name="tools/manifest" path="manifest"></project>
  <project name="platform/manifest" path="platform-manifest"></project>
</manifest>`

	actual, err := xml.MarshalIndent(&m, "", "  ")
	assert.Nil(err)
	assert.Equal(expected, string(actual))
}

func TestMarshal2(t *testing.T) {
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
			Remote:   "origin",
			Revision: "master",
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

func TestUnmarshal2(t *testing.T) {
	assert := assert.New(t)

	m := Manifest{}
	err := xml.Unmarshal(xmlData, &m)

	assert.Nil(err)
	assert.Equal(manifest.Default, m.Default)
	assert.Equal(manifest.Remotes, m.Remotes)
	assert.Equal(manifest.Projects, m.Projects)
}
