package manifest

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Macros for manifest
const (
	ManifestXMLFile = "manifest.xml"
)

// Manifest is for toplevel XML structure
type Manifest struct {
	XMLName        xml.Name        `xml:"manifest"`
	Notice         string          `xml:"notice,omitempty"`
	Remotes        []Remote        `xml:"remote,omitempty"`
	Default        *Default        `xml:"default,omitempty"`
	Server         *Server         `xml:"manifest-server,omitempty"`
	Projects       []Project       `xml:"project,omitempty"`
	RemoveProjects []RemoveProject `xml:"remove-project,omitempty"`
	ExtendProjects []ExtendProject `xml:"extend-project,omitempty"`
	RepoHooks      *RepoHooks      `xml:"repo-hooks,omitempty"`
	Includes       []Include       `xml:"include,omitempty"`
	SourceFile     string          `xml:"-"`
}

// Remote is for remote XML element
type Remote struct {
	Name     string `xml:"name,attr,omitempty"`
	Alias    string `xml:"alias,attr,omitempty"`
	Fetch    string `xml:"fetch,attr,omitempty"`
	PushURL  string `xml:"pushurl,attr,omitempty"`
	Review   string `xml:"review,attr,omitempty"`
	Revision string `xml:"revision,attr,omitempty"`
}

// Default is for default XML element
type Default struct {
	Remote     string `xml:"remote,attr,omitempty"`
	Revision   string `xml:"revision,attr,omitempty"`
	DestBranch string `xml:"dest-branch,attr,omitempty"`
	Upstream   string `xml:"upstream,attr,omitempty"`
	SyncJ      string `xml:"sync-j,attr,omitempty"`
	SyncC      string `xml:"sync-c,attr,omitempty"`
	SyncS      string `xml:"sync-s,attr,omitempty"`
	SyncTags   string `xml:"sync-tags,attr,omitempty"`
}

// Server is for manifest-server XML element
type Server struct {
	URL string `xml:"url,attr,omitempty"`
}

// Project is for project XML element
type Project struct {
	Annotations []Annotation `xml:"annotation,omitempty"`
	Projects    []Project    `xml:"project,omitempty"`
	CopyFiles   []CopyFile   `xml:"copyfile,omitempty"`
	LinkFiles   []LinkFile   `xml:"linkfile,omitempty"`

	Name       string `xml:"name,attr,omitempty"`
	Path       string `xml:"path,attr,omitempty"`
	Remote     string `xml:"remote,attr,omitempty"`
	Revision   string `xml:"revision,attr,omitempty"`
	DestBranch string `xml:"dest-branch,attr,omitempty"`
	Groups     string `xml:"groups,attr,omitempty"`
	SyncC      string `xml:"sync-c,attr,omitempty"`
	SyncS      string `xml:"sync-s,attr,omitempty"`
	SyncTags   string `xml:"sync-tags,attr,omitempty"`
	Upstream   string `xml:"upstream,attr,omitempty"`
	CloneDepth string `xml:"clone-depth,attr,omitempty"`
	ForcePath  string `xml:"force-path,attr,omitempty"`
}

// Annotation is for annotation XML element
type Annotation struct {
	Name  string `xml:"name,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
	Keep  string `xml:"keep,attr,omitempty"`
}

// CopyFile is for copyfile XML element
type CopyFile struct {
	Src  string `xml:"src,attr,omitempty"`
	Dest string `xml:"dest,attr,omitempty"`
}

// LinkFile is for linkfile XML element
type LinkFile struct {
	Src  string `xml:"src,attr,omitempty"`
	Dest string `xml:"dest,attr,omitempty"`
}

// ExtendProject is for extend-project XML element
type ExtendProject struct {
	Name     string `xml:"name,attr,omitempty"`
	Path     string `xml:"path,attr,omitempty"`
	Groups   string `xml:"groups,attr,omitempty"`
	Revision string `xml:"revision,attr,omitempty"`
}

// RemoveProject is for remove-project XML element
type RemoveProject struct {
	Name string `xml:"name,attr,omitempty"`
}

// RepoHooks is for repo-hooks XML element
type RepoHooks struct {
	InProject   string `xml:"in-project,attr,omitempty"`
	EnabledList string `xml:"enabled-list,attr,omitempty"`
}

// Include is for include XML element
type Include struct {
	Name string `xml:"name,attr,omitempty"`
}

func unmarshal(file string) (*Manifest, error) {
	manifest := Manifest{}
	if _, err := os.Stat(file); err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read manifest file '%s': %s", file, err)
	}

	err = xml.Unmarshal(buf, &manifest)
	if err != nil {
		return nil, fmt.Errorf("fail to parse manifest file '%s': %s", file, err)
	}
	return &manifest, nil
}

func parseXML(file string) (*Manifest, error) {
	m, err := unmarshal(file)
	if err != nil {
		return nil, err
	}
	return m, err
}

// Load will load and parse manifest XML file
func Load(repoDir string) (*Manifest, error) {
	var (
		file string
	)

	// Ignore uninitialized repo
	file = filepath.Join(repoDir, ManifestXMLFile)
	if _, err := os.Stat(file); err != nil {
		return nil, nil
	}

	return parseXML(file)
}
