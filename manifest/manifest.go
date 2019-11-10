// Copyright © 2019 Alibaba Co. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package manifest implements marshal and unmarshal of manifest XML.
package manifest

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
)

// Macros for manifest
const (
	maxRecursiveDepth = 10
)

// Manifest is for toplevel XML structure.
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

// Remote is for remote XML element.
type Remote struct {
	Name     string `xml:"name,attr,omitempty"`
	Alias    string `xml:"alias,attr,omitempty"`
	Fetch    string `xml:"fetch,attr,omitempty"`
	PushURL  string `xml:"pushurl,attr,omitempty"`
	Review   string `xml:"review,attr,omitempty"`
	Revision string `xml:"revision,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
}

// Default is for default XML element.
type Default struct {
	RemoteName string `xml:"remote,attr,omitempty"`
	Revision   string `xml:"revision,attr,omitempty"`
	DestBranch string `xml:"dest-branch,attr,omitempty"`
	Upstream   string `xml:"upstream,attr,omitempty"`
	SyncJ      int    `xml:"sync-j,attr,omitempty"`
	SyncC      string `xml:"sync-c,attr,omitempty"`
	SyncS      string `xml:"sync-s,attr,omitempty"`
	SyncTags   string `xml:"sync-tags,attr,omitempty"`
}

// Server is for manifest-server XML element.
type Server struct {
	URL string `xml:"url,attr,omitempty"`
}

// Project is for project XML element.
type Project struct {
	Annotations []Annotation `xml:"annotation,omitempty"`
	Projects    []Project    `xml:"project,omitempty"`
	CopyFiles   []CopyFile   `xml:"copyfile,omitempty"`
	LinkFiles   []LinkFile   `xml:"linkfile,omitempty"`

	Name       string `xml:"name,attr,omitempty"`
	Path       string `xml:"path,attr,omitempty"`
	RemoteName string `xml:"remote,attr,omitempty"`
	Revision   string `xml:"revision,attr,omitempty"`
	DestBranch string `xml:"dest-branch,attr,omitempty"`
	Groups     string `xml:"groups,attr,omitempty"`
	Rebase     string `xml:"rebase,attr,omitempty"`
	SyncC      string `xml:"sync-c,attr,omitempty"`
	SyncS      string `xml:"sync-s,attr,omitempty"`
	SyncTags   string `xml:"sync-tags,attr,omitempty"`
	Upstream   string `xml:"upstream,attr,omitempty"`
	CloneDepth string `xml:"clone-depth,attr,omitempty"`
	ForcePath  string `xml:"force-path,attr,omitempty"`

	isMetaProject  bool    `xml:"-"`
	ManifestRemote *Remote `xml:"-"`
}

// Annotation is for annotation XML element.
type Annotation struct {
	Name  string `xml:"name,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
	Keep  string `xml:"keep,attr,omitempty"`
}

// CopyFile is for copyfile XML element.
type CopyFile struct {
	Src  string `xml:"src,attr,omitempty"`
	Dest string `xml:"dest,attr,omitempty"`
}

// LinkFile is for linkfile XML element.
type LinkFile struct {
	Src  string `xml:"src,attr,omitempty"`
	Dest string `xml:"dest,attr,omitempty"`
}

// ExtendProject is for extend-project XML element.
type ExtendProject struct {
	Name     string `xml:"name,attr,omitempty"`
	Path     string `xml:"path,attr,omitempty"`
	Groups   string `xml:"groups,attr,omitempty"`
	Revision string `xml:"revision,attr,omitempty"`
}

// RemoveProject is for remove-project XML element.
type RemoveProject struct {
	Name string `xml:"name,attr,omitempty"`
}

// RepoHooks is for repo-hooks XML element.
type RepoHooks struct {
	InProject   string `xml:"in-project,attr,omitempty"`
	EnabledList string `xml:"enabled-list,attr,omitempty"`
}

// Include is for include XML element.
type Include struct {
	Name string `xml:"name,attr,omitempty"`
}

// AllProjects returns all projects (include current project and all sub-projects)
// of a project recursively.
func (v *Project) AllProjects(parent *Project) []Project {
	var project Project

	projects := []Project{}
	if parent != nil {
		if parent.Path != "" {
			v.Path = filepath.Join(parent.Path, v.Path)
		}

		if parent.Name != "" {
			v.Name = filepath.Join(parent.Name, v.Name)
		}
	}

	if strings.HasSuffix(v.Name, ".git") {
		v.Name = strings.TrimSuffix(v.Name, ".git")
	}
	v.Name = filepath.Clean(filepath.ToSlash(v.Name))
	v.Path = filepath.Clean(filepath.ToSlash(v.Path))

	// remove field: Projects
	if len(v.Projects) > 0 {
		project = Project{
			Annotations: v.Annotations,
			CopyFiles:   v.CopyFiles,
			LinkFiles:   v.LinkFiles,
			Name:        v.Name,
			Path:        v.Path,
			RemoteName:  v.RemoteName,
			Revision:    v.Revision,
			DestBranch:  v.DestBranch,
			Groups:      v.Groups,
			Rebase:      v.Rebase,
			SyncC:       v.SyncC,
			SyncS:       v.SyncS,
			SyncTags:    v.SyncTags,
			Upstream:    v.Upstream,
			CloneDepth:  v.CloneDepth,
			ForcePath:   v.ForcePath,
		}
		projects = append(projects, project)
	} else {
		projects = append(projects, *v)
	}

	for _, p := range v.Projects {
		projects = append(projects, p.AllProjects(v)...)
	}
	return projects
}

func isTrue(value string, def bool) bool {
	if value == "" {
		return def
	}
	value = strings.ToLower(value)
	if value == "true" ||
		value == "yes" ||
		value == "1" ||
		value == "t" ||
		value == "y" ||
		value == "on" {
		return true
	}
	return false
}

// IsRebase indicates a project should use rebase instead of reset for syncing.
func (v Project) IsRebase() bool {
	return isTrue(v.Rebase, true)
}

// IsSyncS indicates a project should sync submodules.
func (v Project) IsSyncS() bool {
	return isTrue(v.SyncS, false)
}

// IsSyncC indicates a project should sync current branch.
func (v Project) IsSyncC() bool {
	return isTrue(v.SyncC, false)
}

// IsSyncTags indicates a project should sync tags.
func (v Project) IsSyncTags() bool {
	return isTrue(v.SyncTags, true)
}

// IsMetaProject indicates current project is a ManifestProject or not.
func (v Project) IsMetaProject() bool {
	return v.isMetaProject
}

func (v *Manifest) allProjects() []Project {
	projects := []Project{}
	for _, p := range v.Projects {
		projects = append(projects, p.AllProjects(nil)...)
	}
	return projects
}

// AllProjects returns all projects and fill missing fields
func (v *Manifest) AllProjects() []Project {
	projects := v.allProjects()
	remotes := make(map[string]*Remote)
	for i := range v.Remotes {
		remotes[v.Remotes[i].Name] = &v.Remotes[i]
	}

	for i := range projects {
		if projects[i].RemoteName == "" {
			if v.Default == nil || v.Default.RemoteName == "" {
				log.Fatalf("no remote defined for for project '%s'",
					projects[i].Name)
			}
			projects[i].RemoteName = v.Default.RemoteName
		}
		projects[i].ManifestRemote = remotes[projects[i].RemoteName]
		if projects[i].ManifestRemote == nil {
			log.Fatalf("cannot find remote '%s' for project '%s'",
				projects[i].RemoteName,
				projects[i].Name)
		}

		if projects[i].Revision == "" {
			projects[i].Revision = projects[i].ManifestRemote.Revision
		}

		if v.Default != nil {
			if projects[i].Revision == "" {
				projects[i].Revision = v.Default.Revision
			}
			if projects[i].DestBranch == "" {
				projects[i].DestBranch = v.Default.DestBranch
			}
			if projects[i].Upstream == "" {
				projects[i].Upstream = v.Default.Upstream
			}
			if projects[i].SyncC == "" {
				projects[i].SyncC = v.Default.SyncC
			}
			if projects[i].SyncS == "" {
				projects[i].SyncS = v.Default.SyncS
			}
			if projects[i].SyncTags == "" {
				projects[i].SyncTags = v.Default.SyncTags
			}
		}

		if projects[i].Revision == "" {
			log.Fatalf("no revision for project '%s'", projects[i].Name)
		}
	}
	return projects
}

// Merge implements merging another manifest to self.
func (v *Manifest) Merge(m *Manifest) error {
	if m.Notice != "" {
		if v.Notice == "" {
			v.Notice = m.Notice
		} else {
			return fmt.Errorf("duplicate notice in %s", m.SourceFile)
		}
	}

	for _, r1 := range m.Remotes {
		found := false
		for _, r2 := range v.Remotes {
			if r1.Name == r2.Name {
				if r1 != r2 {
					return fmt.Errorf("duplicate remote in %s", m.SourceFile)
				}
				found = true
				break
			}
		}
		if !found {
			v.Remotes = append(v.Remotes, r1)
		}
	}

	if m.Default != nil {
		if v.Default != nil {
			if v.Default != m.Default {
				return fmt.Errorf("duplicate default in %s", m.SourceFile)
			}
		} else {
			v.Default = m.Default
		}
	}

	if m.Server != nil {
		if v.Server != nil {
			if v.Server != m.Server {
				return fmt.Errorf("duplicate manifest-server in %s", m.SourceFile)
			}
		} else {
			v.Server = m.Server
		}
	}

	realPath := make(map[string]bool)
	for _, p := range v.allProjects() {
		if realPath[p.Path] {
			return fmt.Errorf("duplicate path for project '%s' in '%s'",
				p.Path,
				v.SourceFile)
		}
		realPath[p.Path] = true
	}
	for _, p := range m.allProjects() {
		p.Name = cleanPath(p.Name)
		p.Path = cleanPath(p.Path)
		if realPath[p.Path] {
			return fmt.Errorf("duplicate path for project '%s' in '%s'",
				p.Path,
				m.SourceFile)
		}
		v.Projects = append(v.Projects, p)
		realPath[p.Path] = true
	}

	rmPath := make(map[string]bool)
	for _, r := range m.RemoveProjects {
		r.Name = cleanPath(r.Name)
		rmPath[r.Name] = true
	}
	ps := []Project{}
	for _, p := range v.allProjects() {
		if rmPath[p.Name] {
			realPath[p.Path] = false
		} else {
			ps = append(ps, p)
		}
	}
	v.Projects = ps

	extPath := make(map[string]ExtendProject)
	for _, p := range m.ExtendProjects {
		p.Name = cleanPath(p.Name)
		extPath[p.Name] = p
	}
	for i, p := range v.Projects {
		if p2, ok := extPath[p.Name]; ok {
			if p2.Path == p.Path {
				if p.Groups == "" {
					v.Projects[i].Groups = p2.Groups
				} else if p2.Groups != "" {
					groups := []string{}
					groups = append(groups, strings.Split(p.Groups, ",")...)
					groups = append(groups, strings.Split(p2.Groups, ",")...)
					v.Projects[i].Groups = strings.Join(groups, ",")
				}
				if p2.Revision != "" {
					v.Projects[i].Revision = p2.Revision
				}
			}
		}
	}

	// m.RepoHooks

	return nil
}

// ProjectHandler is an interface to manipulate projects of manifest
type ProjectHandler interface {
	// The 1st parameter is pointer of a project, and the 2nd parameter
	// is parent dir of the path of the project.
	Process(*Project, string) error
}

// ProjectHandle executes Process method of the given handle on each project
func (v *Manifest) ProjectHandle(handle ProjectHandler) error {
	for i := range v.Projects {
		err := v.Projects[i].execute(handle, "", 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *Project) execute(handle ProjectHandler, path string, depth int) error {
	if depth > maxRecursiveDepth {
		return fmt.Errorf("recursive is to deep")
	}

	err := handle.Process(v, path)
	if err != nil {
		return err
	}

	if path == "" {
		path = v.Path
	} else {
		path = filepath.Join(path, v.Path)
	}

	for i := range v.Projects {
		err := v.Projects[i].execute(handle, path, depth+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanPath(name string) string {
	return filepath.Clean(strings.Replace(strings.TrimSuffix(name, ".git"), "\\", "/", -1))
}

func unmarshalFile(file string) (*Manifest, error) {
	if _, err := os.Stat(file); err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read manifest file '%s': %s", file, err)
	}

	ms, err := Unmarshal(buf)
	if err != nil {
		return nil, fmt.Errorf("fail to parse manifest file '%s': %s", file, err)
	}

	return ms, nil
}

func parseXML(file string, depth int) ([]*Manifest, error) {
	ms := []*Manifest{}

	m, err := unmarshalFile(file)
	if err != nil {
		return ms, err
	}
	if m == nil {
		return ms, nil
	}
	m.SourceFile = file
	ms = append(ms, m)

	for _, i := range m.Includes {
		f, err := path.AbsJoin(filepath.Dir(file), i.Name)
		if err != nil {
			return ms, err
		}

		if depth > maxRecursiveDepth {
			return nil, fmt.Errorf("exceeded maximum include depth (%d) while including\n"+
				"\t%s\n"+
				"from"+
				"\t%s\n"+
				"This might be due to circular includes",
				maxRecursiveDepth,
				f,
				file)
		}

		subMs, err := parseXML(f, depth+1)
		if err != nil {
			return ms, err
		}
		ms = append(ms, subMs...)
	}

	return ms, nil
}

func mergeManifests(ms []*Manifest) (*Manifest, error) {
	manifest := &Manifest{}
	for _, m := range ms {
		err := manifest.Merge(m)
		if err != nil {
			return nil, err
		}
	}
	return manifest, nil
}

// Load implements load and parse manifest XML file in repoDir.
func Load(repoDir string) (*Manifest, error) {
	var (
		file string
		err  error
	)

	file = filepath.Join(repoDir, config.ManifestXML)
	if _, err = os.Stat(file); err != nil {
		defaultXML := ""
		manifestsDir := filepath.Join(repoDir, config.Manifests)
		cfg, err := goconfig.Load(manifestsDir)
		if err != nil && err != goconfig.ErrNotExist {
			return nil, fmt.Errorf("fail to read config from %s: %s", manifestsDir, err)
		}
		if cfg != nil {
			defaultXML = cfg.Get(config.CfgManifestName)
		}
		if defaultXML == "" {
			defaultXML = config.DefaultXML
		}
		file = filepath.Join(manifestsDir, defaultXML)
		if _, err = os.Stat(file); err != nil {
			return nil, err
		}
	}
	return LoadFile(repoDir, file)
}

// LoadFile implements load specific manifest file inside repoDir.
func LoadFile(repoDir, file string) (*Manifest, error) {
	var (
		dir       string
		err       error
		manifests = []*Manifest{}
	)

	if !filepath.IsAbs(file) {
		file = filepath.Join(repoDir, config.Manifests, file)
	}

	// Ignore uninitialized repo
	if _, err := os.Stat(file); err != nil {
		return nil, nil
	}

	ms, err := parseXML(file, 1)
	if err != nil {
		return nil, err
	}
	manifests = append(manifests, ms...)

	// load local_manifest.xml (obsolete)
	files := []string{}
	file = filepath.Join(repoDir, config.LocalManifestXML)
	dir = filepath.Join(repoDir, config.LocalManifests)
	if _, err = os.Stat(file); err == nil {
		log.Warnf("%s is deprecated; put local manifests in `%s` instead", file, dir)
		files = append(files, file)
	}

	// load xml files in local_manifests
	if _, err = os.Stat(dir); err == nil {
		filepath.Walk(dir, func(name string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if dir == name {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(name, ".xml") {
				files = append(files, name)
			}
			return nil
		})
	}

	for _, file = range files {
		ms, err := parseXML(file, 1)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, ms...)
	}

	return mergeManifests(manifests)
}

// Unmarshal implements decoding XML (in buf) to manifest.
func Unmarshal(buf []byte) (*Manifest, error) {
	var ms = Manifest{}

	err := xml.Unmarshal(buf, &ms)
	return &ms, err
}

// Marshal implements encoding manifest to XML.
func Marshal(ms *Manifest) ([]byte, error) {
	return xml.MarshalIndent(ms, "", "  ")
}

// ManifestsProject is a special instance of Project.
var ManifestsProject = &Project{
	Name:          "manifests",
	Path:          "manifests",
	RemoteName:    "origin",
	Revision:      "refs/heads/master",
	DestBranch:    "master",
	isMetaProject: true,
}
