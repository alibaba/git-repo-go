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

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/aliyun/git-repo-go/cap"
	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/file"
	"github.com/aliyun/git-repo-go/project"
	"github.com/aliyun/git-repo-go/workspace"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

const (
	// syncDefaultJobs is the default value of --jobs
	syncDefaultJobs = 4
)

type syncCommand struct {
	WorkSpaceCommand

	cmd          *cobra.Command
	FetchOptions project.FetchOptions

	O struct {
		ForceBroken            bool
		ForceSync              bool
		LocalOnly              bool
		NetworkOnly            bool
		DetachHead             bool
		CurrentBranchOnly      bool
		Jobs                   int
		ManifestName           string
		NoCache                bool
		NoCloneBundle          bool
		ManifestServerUsername string
		ManifestServerPassword string
		FetchSubmodules        bool
		NoTags                 bool
		OptimizedFetch         bool
		Prune                  bool
		SmartSync              bool
		SmartTag               string
	}
}

func (v *syncCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "sync",
		Short: "Update working tree to the latest revision",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}
	v.cmd.Flags().BoolVarP(&v.O.ForceBroken,
		"force-broken",
		"f",
		false,
		"continue sync even if a project fails to sync")
	v.cmd.Flags().BoolVar(&v.O.ForceSync,
		"force-sync",
		false,
		"overwrite an existing git directory if it needs to "+
			"point to a different object directory. WARNING: this "+
			"may cause loss of data")
	v.cmd.Flags().BoolVarP(&v.O.LocalOnly,
		"local-only",
		"l",
		false,
		"only update working tree, don't fetch")
	v.cmd.Flags().BoolVarP(&v.O.NetworkOnly,
		"network-only",
		"n",
		false,
		"fetch only, don't update working tree")
	v.cmd.Flags().BoolVarP(&v.O.DetachHead,
		"detach",
		"d",
		false,
		"detach projects back to manifest revision")
	v.cmd.Flags().BoolVarP(&v.O.CurrentBranchOnly,
		"current-branch",
		"c",
		false,
		"fetch only current branch from server")
	v.cmd.Flags().IntVarP(&v.O.Jobs,
		"jobs",
		"j",
		v.manifestsDefaultJobs(),
		fmt.Sprintf("projects to fetch simultaneously"))
	v.cmd.Flags().StringVarP(&v.O.ManifestName,
		"manifest-name",
		"m",
		"",
		"temporary manifest to use for this sync")
	v.cmd.Flags().BoolVar(&v.O.NoCache,
		"no-cache",
		false,
		"Ignore ssh-info cache, and recheck ssh-info API")
	v.cmd.Flags().BoolVar(&v.O.NoCloneBundle,
		"no-clone-bundle",
		false,
		"disable use of /clone.bundle on HTTP/HTTPS")
	v.cmd.Flags().StringVarP(&v.O.ManifestServerUsername,
		"manifest-server-username",
		"u",
		"",
		"username to authenticate with the manifest server")
	v.cmd.Flags().StringVarP(&v.O.ManifestServerPassword,
		"manifest-server-password",
		"p",
		"",
		"password to authenticate with the manifest server")
	v.cmd.Flags().BoolVar(&v.O.FetchSubmodules,
		"fetch-submodules",
		false,
		"fetch submodules from server")
	v.cmd.Flags().BoolVar(&v.O.NoTags,
		"no-tags",
		false,
		"don't fetch tags")
	v.cmd.Flags().BoolVar(&v.O.OptimizedFetch,
		"optimized-fetch",
		false,
		"only fetch projects fixed to sha1 if revision does not exist locally")
	v.cmd.Flags().BoolVar(&v.O.Prune,
		"prune",
		false,
		"delete refs that no longer exist on the remote")
	v.cmd.Flags().BoolVar(&v.O.SmartSync,
		"smart-sync",
		false,
		"smart sync using manifest from the latest known good build")
	v.cmd.Flags().StringVarP(&v.O.SmartTag,
		"smart-tag",
		"t",
		"",
		"smart sync using manifest from a known tag")

	return v.cmd
}

// Value of manifestsDefaultJobs() is used as command arg's default value.
// Do not fail if run git-repo command out of a workspace.
func (v *syncCommand) manifestsDefaultJobs() int {
	var nJobs int

	rws, _ := workspace.NewRepoWorkSpace("")
	if rws != nil &&
		rws.Manifest != nil &&
		rws.Manifest.Default != nil &&
		rws.Manifest.Default.SyncJ > 0 {
		nJobs = rws.Manifest.Default.SyncJ
	}

	if nJobs <= 0 {
		nJobs = syncDefaultJobs
	}

	return nJobs
}

func (v *syncCommand) maxSyncJobs() int {
	var (
		nJobs int = config.MaxJobs
	)

	noFile, err := cap.GetRlimitNoFile()
	if err == nil {
		nJobs = min(int((noFile-5)/3), config.MaxJobs)
	}

	if nJobs <= 0 {
		nJobs = 1
	}

	return nJobs
}

func (v syncCommand) CallManifestServerRPC() {
	// TODO: implement `_SmartSyncSetup`
	log.Panic("not implement CallManifestServerRPC")
}

func (v *syncCommand) updateManifestProject() error {
	var err error

	ws := v.RepoWorkSpace()
	mp := ws.ManifestProject
	s := mp.ReadSettings()
	track := mp.TrackBranch("")

	if track == "" {
		log.Notef("manifest project is not updated, for there is no tracking branch")
		return nil
	}

	if !v.O.LocalOnly {
		// Fetch repositories
		fetchOptions := project.FetchOptions{
			RepoSettings: *s,

			CurrentBranchOnly: v.O.CurrentBranchOnly,
			NoTags:            v.O.NoTags,
			OptimizedFetch:    v.O.OptimizedFetch,
			Quiet:             config.GetQuiet(),
		}

		err = mp.SyncNetworkHalf(&fetchOptions)
		if err != nil {
			return err
		}
	}

	// Get current manifest version
	oldrev, _ := mp.ResolveRevision("HEAD")

	// No update found in manifest project
	newrev, _ := mp.ResolveRemoteTracking(track)
	if oldrev == newrev {
		return nil
	}

	// Has commit not yet checkout?
	revlist, err := mp.Revlist(newrev, "--not", oldrev)
	if err != nil {
		return err
	}
	if len(revlist) == 0 {
		return nil
	}

	// Checkout
	checkoutOptions := project.CheckoutOptions{
		RepoSettings: *s,

		Quiet: config.GetQuiet(),
	}
	err = mp.SyncLocalHalf(&checkoutOptions)
	if err != nil {
		return err
	}

	// Reload Manifest
	v.ReloadRepoWorkSpace()

	// Load different manifest file
	if v.O.ManifestName != "" {
		v.RepoWorkSpace().Override(v.O.ManifestName)
	}

	return nil
}

func (v syncCommand) NetworkHalf(allProjects []*project.Project) error {
	var (
		err  error
		errs []error
	)

	jobs := v.O.Jobs
	if jobs < 1 {
		jobs = 1
	}

	// TODO 1. Record fetch time, save time and project name to JSON
	// TODO 2. Sort projects by its fetch time (reverse order).

	projectsByName := project.IndexByName(allProjects)
	jobTasks := make(chan string, jobs)
	jobResults := make(chan error, jobs)

	worker := func(i int) {
		var (
			err      error
			name     string
			projects []*project.Project
			p        *project.Project
		)

		log.Debugf("start NetworkHalf worker #%d", i)
		for name = range jobTasks {
			projects = projectsByName[name]
			for _, p = range projects {
				log.Debugf("worker #%d: sync %s", i, p.Name)
				err = p.SyncNetworkHalf(&v.FetchOptions)
				jobResults <- err
			}
		}
	}

	for i := 0; i < jobs; i++ {
		go worker(i)
	}

	go func() {
		for name := range projectsByName {
			jobTasks <- name
		}

		close(jobTasks)
	}()

	for i := 0; i < len(projectsByName); i++ {
		err = <-jobResults
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	errMsg := ""
	for _, err = range errs {
		errMsg += err.Error() + "\n"
	}
	return errors.New(errMsg)
}

func (v syncCommand) LocalHalf(allProjects []*project.Project) error {
	var (
		err  error
		errs []error
		wg   sync.WaitGroup
	)

	jobs := v.O.Jobs
	if jobs < 1 {
		jobs = 1
	}

	jobTasks := make(chan *project.Tree, jobs)

	checkoutOptions := project.CheckoutOptions{
		Quiet:      config.GetQuiet(),
		DetachHead: v.O.DetachHead,
	}

	wg.Add(len(allProjects))

	worker := func(i int) {
		var (
			err  error
			tree *project.Tree
			p    *project.Project
		)

		log.Debugf("start LocalHalf worker #%d", i)
		for tree = range jobTasks {
			p = tree.Project
			if p != nil {
				log.Debugf("worker #%d: checkout %s", i, p.Name)
				err = p.SyncLocalHalf(&checkoutOptions)
				if err != nil {
					errs = append(errs, err)
				}
			}

			go func(tree project.Tree) {
				for _, t := range tree.Trees {
					jobTasks <- t
				}
			}(*tree)

			// if p is nil, it's root tree
			if p != nil {
				log.Debugf("worker #%d: done %s", i, p.Name)
				wg.Done()
			}
		}
	}

	for i := 0; i < jobs; i++ {
		go worker(i)
	}

	tree := project.ProjectsTree(allProjects)
	jobTasks <- tree

	wg.Wait()
	close(jobTasks)

	if len(errs) == 0 {
		return nil
	}

	errMsg := ""
	for _, err = range errs {
		errMsg += err.Error() + "\n"
	}
	return errors.New(errMsg)
}

// findObsoletePaths returns obsolete paths.
// Please note that the oldPaths and newPaths must be sorted.
func (v syncCommand) findObsoletePaths(oldPaths, newPaths []string) []string {
	result := []string{}

	i, j := 0, 0
	for i < len(oldPaths) && j < len(newPaths) {
		if oldPaths[i] < newPaths[j] {
			result = append(result, oldPaths[i])
			i++
		} else if oldPaths[i] > newPaths[j] {
			j++
		} else {
			i++
			j++
		}
	}
	for i < len(oldPaths) {
		result = append(result, oldPaths[i])
		i++
	}

	return result
}

func (v syncCommand) findGitWorktree(dir string) []string {
	result := []string{}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, err = os.Stat(filepath.Join(path, ".git")); err == nil {
				result = append(result, path)
				return filepath.SkipDir
			}
		}
		return nil
	})

	return result
}

func (v syncCommand) removeEmptyDirs(dir string) {
	var (
		oldDir = dir
		rws    = v.RepoWorkSpace()
	)

	for {
		if !strings.HasPrefix(dir, rws.RootDir) {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
		if dir == oldDir {
			break
		}
		oldDir = dir
	}
}

func (v syncCommand) removeWorktree(dir string, gitTrees []string) error {
	var err error

	if len(gitTrees) == 0 {
		log.Printf("will remove %s", dir)
		return os.RemoveAll(dir)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Printf("will remove %s", path)
			err = os.Remove(path)
			if err != nil {
				return err
			}
			return nil
		}
		for _, p := range gitTrees {
			if path == p {
				return filepath.SkipDir
			}
			if strings.HasPrefix(p, path+string(os.PathSeparator)) || strings.HasPrefix(p, path+"/") {
				return nil
			}
		}
		log.Printf("will remove all %s", path)
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
		return filepath.SkipDir
	})

	return err
}

func (v syncCommand) removeObsoletePaths(oldPaths, newPaths []string) error {
	var err error

	sort.Strings(oldPaths)
	sort.Strings(newPaths)
	obsoletePaths := v.findObsoletePaths(oldPaths, newPaths)

	ws := v.RepoWorkSpace()

	for _, p := range obsoletePaths {
		workdir := filepath.Clean(filepath.Join(ws.RootDir, p))
		gitdir := filepath.Join(workdir, ".git")
		workRepoPath := filepath.Clean(filepath.Join(ws.RootDir, config.DotRepo, p+".git"))

		if !strings.HasPrefix(workdir, ws.RootDir) {
			return fmt.Errorf("cannot delete project path '%s', which beyond repo root '%s'", workdir, ws.RootDir)
		}

		if _, err := os.Stat(gitdir); err != nil {
			log.Debugf("cannot find gitdir '%s' when removing obsolete path", gitdir)
			continue
		}

		if ok, _ := project.IsClean(workdir); !ok {
			return fmt.Errorf(`Cannot remove project "%s": uncommitted changes are present.
Please commit changes, then run sync again`,
				p)
		}

		// Remove gitdir first
		err = os.RemoveAll(gitdir)
		if err != nil {
			return fmt.Errorf("fail to remove '%s': %s",
				gitdir,
				err)
		}

		// Remove worktree, except recursive git repositories
		ignoreRepos := v.findGitWorktree(workdir)
		err = v.removeWorktree(workdir, ignoreRepos)
		if err != nil {
			return fmt.Errorf("fail to remove '%s': %s", workdir, err)
		}
		v.removeEmptyDirs(workdir)

		// Remove project repository
		if _, err = os.Stat(workRepoPath); err != nil {
			if !strings.HasPrefix(workRepoPath, ws.RootDir) {
				return fmt.Errorf("cannot delete project repo '%s', which beyond repo root '%s'", workRepoPath, ws.RootDir)
			}
			err = os.RemoveAll(workRepoPath)
			if err != nil {
				return err
			}
			v.removeEmptyDirs(workRepoPath)
		}
	}

	return nil
}

func (v syncCommand) UpdateProjectList() error {
	var (
		newPaths = []string{}
		oldPaths = []string{}
		ws       = v.RepoWorkSpace()
	)

	allProjects, err := ws.GetProjects(&workspace.GetProjectsOptions{
		MissingOK:    true,
		SubmodulesOK: v.O.FetchSubmodules,
	})
	if err != nil {
		return err
	}

	for _, p := range allProjects {
		newPaths = append(newPaths, p.Path)
	}

	projectListFile := filepath.Join(ws.RootDir, config.DotRepo, "project.list")
	if _, err = os.Stat(projectListFile); err == nil {
		f, err := os.Open(projectListFile)
		defer f.Close()
		if err != nil {
			log.Fatalf("fail to open %s: %s", projectListFile, err)
		}
		r := bufio.NewReader(f)
		for {
			line, err := r.ReadString('\n')
			line = strings.TrimSpace(line)
			if line != "" {
				oldPaths = append(oldPaths, line)
			}
			if err != nil {
				break
			}
		}
	}

	err = v.removeObsoletePaths(oldPaths, newPaths)
	if err != nil {
		return err
	}

	projectListLockFile := projectListFile + ".lock"
	lockf, err := file.New(projectListLockFile).OpenCreateRewriteExcl()
	if err != nil {
		return fmt.Errorf("fail to create lockfile '%s': %s", projectListLockFile, err)
	}
	defer lockf.Close()
	for _, p := range newPaths {
		_, err = lockf.WriteString(p + "\n")
		if err != nil {
			return fmt.Errorf("fail to save lockfile '%s': %s", projectListLockFile, err)
		}
	}
	lockf.Close()

	err = os.Rename(projectListLockFile, projectListFile)
	if err != nil {
		return fmt.Errorf("fail to rename lockfile to '%s': %s", projectListFile, err)
	}

	return nil
}

func (v syncCommand) Execute(args []string) error {
	var (
		err error
	)

	rws := v.RepoWorkSpace()

	if v.O.Jobs > 0 {
		v.O.Jobs = min(v.O.Jobs, v.maxSyncJobs())
	} else {
		v.O.Jobs = 1
	}
	if v.O.NetworkOnly && v.O.DetachHead {
		return newUserError("cannot combine -n and -d")
	}
	if v.O.NetworkOnly && v.O.LocalOnly {
		return newUserError("cannot combine -n and -l")
	}
	if v.O.ManifestName != "" && v.O.SmartSync {
		return newUserError("cannot combine -m and -s")
	}
	if v.O.ManifestName != "" && v.O.SmartTag != "" {
		return newUserError("cannot combine -m and -t")
	}
	if v.O.ManifestServerUsername != "" || v.O.ManifestServerPassword != "" {
		if !(v.O.SmartSync || v.O.SmartTag != "") {
			return newUserError("-u and -p may only be combined with -s or -t")
		}
		if v.O.ManifestServerUsername == "" || v.O.ManifestServerPassword == "" {
			return newUserError("both -u and -p must be given")
		}
	}

	if v.O.ManifestName != "" {
		rws.Override(v.O.ManifestName)
	}

	v.FetchOptions = project.FetchOptions{
		RepoSettings: *(rws.Settings()),

		Quiet:             config.GetQuiet(),
		CloneBundle:       !v.O.NoCloneBundle,
		CurrentBranchOnly: v.O.CurrentBranchOnly,
		ForceSync:         v.O.ForceSync,
		NoTags:            v.O.NoTags,
		OptimizedFetch:    v.O.OptimizedFetch,
		Prune:             v.O.Prune,
	}

	smartSyncManifestName := "smart_sync_override.xml"
	smartSyncManifestPath := filepath.Join(rws.ManifestProject.WorkDir, smartSyncManifestName)

	if v.O.SmartSync || v.O.SmartTag != "" {
		v.CallManifestServerRPC()
	} else {
		if _, err = os.Stat(smartSyncManifestPath); err == nil {
			err = os.Remove(smartSyncManifestPath)
			if err != nil {
				log.Fatalf("failed to remove existing smart sync override manifest: %s", smartSyncManifestPath)
			}
		}
	}

	err = v.updateManifestProject()
	if err != nil {
		return err
	}

	// Use reloaded WorkSpace after calling `updateManifestProject()`.
	rws = v.RepoWorkSpace()

	allProjects, err := rws.GetProjects(&workspace.GetProjectsOptions{
		Groups:       rws.Settings().Groups,
		MissingOK:    true,
		SubmodulesOK: v.O.FetchSubmodules,
	}, args...)

	if !v.O.LocalOnly {
		err = v.NetworkHalf(allProjects)
		if err != nil {
			return err
		}
	}

	if v.O.NetworkOnly ||
		rws.ManifestProject.MirrorEnabled() ||
		rws.ManifestProject.ArchiveEnabled() {
		return nil
	}

	// Call ssh_info API to detect types of remote servers
	err = rws.LoadRemotes(v.O.NoCache)
	if err != nil {
		log.Notef("fail to check remote server, you may need to install gerrit hooks by hands")
		log.Error(err)
	}

	// Remove obsolete projects
	if err = v.UpdateProjectList(); err != nil {
		log.Fatal(err)
	}

	err = v.LocalHalf(allProjects)
	if err != nil {
		return err
	}

	// If there's a notice that's supposed to print at the end of the sync,
	// print it now...
	if rws.Manifest != nil && rws.Manifest.Notice != "" {
		log.Note(rws.Manifest.Notice)
	}

	return nil
}

var syncCmd = syncCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: true,
		SingleOK: false,
	},
}

func init() {
	rootCmd.AddCommand(syncCmd.Command())
}
