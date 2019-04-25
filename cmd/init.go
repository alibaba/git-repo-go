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
	"fmt"
	"os"
	"runtime"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/color"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/errors"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/project"
	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

type initCommand struct {
	cmd *cobra.Command
	ws  *workspace.WorkSpace

	O struct {
		ManifestURL       string
		ManifestBranch    string
		ManifestName      string
		Archive           bool
		ConfigName        bool
		CurrentBranchOnly bool
		Depth             int
		DetachHead        bool
		Dissociate        bool
		Groups            string
		Mirror            bool
		NoCloneBundle     bool
		NoTags            bool
		Platform          string
		Reference         string
		Submodules        bool
	}
}

func (v *initCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize manifest repo in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runE(args)
		},
	}

	v.cmd.Flags().StringVarP(&v.O.ManifestURL,
		"manifest-url",
		"u",
		"",
		"manifest repository location")
	v.cmd.Flags().StringVarP(&v.O.ManifestBranch,
		"manifest-branch",
		"b",
		"",
		"manifest branch or revision")
	v.cmd.Flags().BoolVar(&v.O.CurrentBranchOnly,
		"current-branch",
		false,
		"fetch only current manifest branch from server")
	v.cmd.Flags().StringVarP(&v.O.ManifestName,
		"manifest-name",
		"m",
		"default.xml",
		"initial manifest file")
	v.cmd.Flags().BoolVarP(&v.O.DetachHead,
		"detach",
		"d",
		false,
		"remove default branch to make manifest project detached")
	v.cmd.Flags().BoolVar(&v.O.Mirror,
		"mirror",
		false,
		"create a replica of the remote repositories rather than a client working directory")
	v.cmd.Flags().StringVar(&v.O.Reference,
		"reference",
		"",
		"location of mirror directory")
	v.cmd.Flags().BoolVar(&v.O.Dissociate,
		"dissociate",
		false,
		"dissociate from reference mirrors after clone")
	v.cmd.Flags().IntVar(&v.O.Depth,
		"depth",
		0,
		"create a shallow clone with given depth; see git clone")
	v.cmd.Flags().BoolVar(&v.O.Archive,
		"archive",
		false,
		"checkout an archive instead of a git repository for each project. See git archive.")
	v.cmd.Flags().BoolVar(&v.O.Submodules,
		"submodules",
		false,
		"sync any submodules associated with the manifest repo")
	v.cmd.Flags().StringVarP(&v.O.Groups,
		"groups",
		"g",
		"default",
		"restrict manifest projects to ones with specified group(s) [default|all|G1,G2,G3|G4,-G5,-G6]")
	v.cmd.Flags().StringVarP(&v.O.Platform,
		"platform",
		"p",
		"auto",
		"restrict manifest projects to ones with a specified platform group [auto|all|none|linux|darwin|...]")
	v.cmd.Flags().BoolVar(&v.O.NoCloneBundle,
		"no-clone-bundle",
		false,
		"disable use of /clone.bundle on HTTP/HTTPS")
	v.cmd.Flags().BoolVar(&v.O.NoTags,
		"no-tags",
		false,
		"don't fetch tags in the manifest")
	v.cmd.Flags().BoolVar(&v.O.ConfigName,
		"config-name",
		false,
		"Always prompt for name/e-mail")

	return v.cmd
}

func (v initCommand) getGroups() string {
	allPlatforms := []string{"linux", "darwin", "windows"}
	groups := []string{}
	for _, g := range strings.Split(v.O.Groups, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			groups = append(groups, g)
		}
	}
	platformize := func(x string) string { return "platform-" + x }
	if v.O.Platform == "auto" {
		isMirror := v.ws.ManifestProject.Config().GetBool("repo.mirror", false)
		if !v.O.Mirror && !isMirror {
			groups = append(groups, platformize(runtime.GOOS))
		}
	} else if v.O.Platform != "" {
		found := false
		for _, sys := range allPlatforms {
			if v.O.Platform == "all" || v.O.Platform == sys {
				groups = append(groups, platformize(sys))
				found = true
			}
		}
		if !found {
			log.Fatalf("invalid platform flag: %s", v.O.Platform)
		}
	}

	groupStr := strings.Join(groups, ",")
	if v.O.Platform == "auto" &&
		groupStr == "default,"+platformize(runtime.GOOS) {
		groupStr = ""
	}

	return groupStr
}

func (v initCommand) runE(args []string) error {
	var (
		err   error
		isNew bool
		ws    *workspace.WorkSpace
	)

	if v.O.Archive && v.O.Mirror {
		log.Fatal("--mirror and --archive cannot be used together")
	}

	if v.O.ManifestURL != "" {
		if strings.HasSuffix(v.O.ManifestURL, "/") {
			v.O.ManifestURL = strings.TrimRight(v.O.ManifestURL, "/")
		}
	}

	// Find repo workspace and load it if exist
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := path.FindRepoRoot(cwd)
	if err != nil {
		if err == errors.ErrRepoDirNotFound {
			repoRoot = cwd
		} else {
			return fmt.Errorf("fail to find .repo: %s", err)
		}
	}

	// Check initialized or not
	if !workspace.Exists(repoRoot) {
		isNew = true
		if v.O.ManifestURL == "" {
			log.Fatal("option --manifest-url (-u) is required")
		}
		ws, err = workspace.NewWorkSpaceInit(repoRoot, v.O.ManifestURL)
		v.ws = ws
		if err != nil {
			return err
		}
		ws.ManifestProject.GitInit()
		// Reload settings
		ws.ManifestProject.ReadSettings()
	} else {
		ws, err = workspace.NewWorkSpace(repoRoot)
		v.ws = ws
		if err != nil {
			return err
		}
		if v.O.ManifestURL != "" && v.O.ManifestURL != ws.ManifestURL() {
			ws.ManifestProject.Settings.ManifestURL = v.O.ManifestURL
			ws.ManifestProject.GitInit()
		}
	}

	if v.O.ManifestBranch != "" {
		v.ws.ManifestProject.SetRevision(v.O.ManifestBranch)
	} else if isNew {
		v.ws.ManifestProject.SetRevision("master")
	} else {
		track := v.ws.ManifestProject.RemoteTrackBranch("")
		if track != "" {
			v.ws.ManifestProject.SetRevision(track)
		} else {
			v.ws.ManifestProject.SetRevision("")
		}
	}

	// Update manifest project settings
	s := v.ws.ManifestProject.Settings
	changed := false
	if v.cmd.Flags().Changed("manifest-url") && s.ManifestURL != v.O.ManifestURL {
		changed = true
		s.ManifestURL = v.O.ManifestURL
	}

	// v.O.ManifestName has default value, and use it if setting is empty
	if (v.cmd.Flags().Changed("manifest-name") && s.ManifestName != v.O.ManifestName) ||
		s.ManifestName == "" {
		changed = true
		s.ManifestName = v.O.ManifestName
	}

	// v.O.Groups has default value, and use it if setting is empty
	groupStr := v.getGroups()
	if ((v.cmd.Flags().Changed("groups") || v.cmd.Flags().Changed("platform")) &&
		s.Groups != groupStr) ||
		s.Groups == "" {
		changed = true
		s.Groups = groupStr
	}

	if v.cmd.Flags().Changed("reference") {
		if v.O.Reference != "" {
			v.O.Reference, err = path.Abs(v.O.Reference)
			if err != nil {
				v.O.Reference = ""
			}
		}
		if s.Reference != v.O.Reference {
			changed = true
			s.Reference = v.O.Reference
		}
	}

	if v.cmd.Flags().Changed("depth") && s.Depth != v.O.Depth {
		changed = true
		s.Depth = v.O.Depth
	}

	if v.cmd.Flags().Changed("archive") && s.Archive != v.O.Archive {
		changed = true
		if !isNew {
			log.Fatal(`--archive is only supported when initializing a new workspace.
Either delete the .repo folder in this workspace, or initialize in another location.`)
		}

	}

	if v.cmd.Flags().Changed("dissociate") && s.Dissociate != v.O.Dissociate {
		changed = true
		s.Dissociate = v.O.Dissociate
	}

	if v.cmd.Flags().Changed("mirror") && s.Mirror != v.O.Mirror {
		changed = true
		s.Mirror = v.O.Mirror
		if !isNew {
			log.Fatal(`--mirror is only supported when initializing a new workspace.
Either delete the .repo folder in this workspace, or initialize in another location.`)
		}

	}

	if v.cmd.Flags().Changed("submodules") && s.Submodules != v.O.Submodules {
		changed = true
		s.Submodules = v.O.Submodules
	}

	if changed {
		err = v.ws.ManifestProject.SaveSettings(s)
		if err != nil {
			return err
		}
	}

	// Fetch repositories
	fetchOptions := project.FetchOptions{
		RepoSettings: *s,

		CloneBundle:       !v.O.NoCloneBundle,
		CurrentBranchOnly: v.O.CurrentBranchOnly,
		ForceSync:         false,
		IsNew:             isNew,
		NoTags:            v.O.NoTags,
		OptimizedFetch:    false,
		Prune:             false,
		Quiet:             config.GetQuiet(),
	}

	err = v.ws.ManifestProject.SyncNetworkHalf(&fetchOptions)
	if err != nil && isNew {
		if !strings.HasPrefix(v.ws.ManifestProject.WorkRepository.Path, v.ws.RootDir) ||
			v.ws.RootDir == "" {
			log.Fatalf("manifest workdir '%s' beyond repo root '%s'", v.ws.ManifestProject.WorkRepository.Path, v.ws.RootDir)
		}
		// Better delete the manifest git dir if we created it; otherwise next
		// time (when user fixes problems) we won't go through the "isNew" logic.
		os.RemoveAll(v.ws.ManifestProject.WorkRepository.Path)
	}

	// Checkout
	checkoutOptions := project.CheckoutOptions{
		RepoSettings: *s,

		Quiet:      config.GetQuiet(),
		DetachHead: false,
	}
	err = v.ws.ManifestProject.SyncLocalHalf(&checkoutOptions)
	if err != nil {
		return err
	}

	if v.O.DetachHead {
		if v.ws.ManifestProject.GetHead() != "" {
			if err = v.ws.ManifestProject.DetachHead(); err != nil {
				return nil
			}

			if err = v.ws.ManifestProject.DeleteBranch("refs/heads/default"); err != nil {
				return nil
			}
		}
	} else if isNew || v.ws.ManifestProject.GetHead() == "" {
		err := v.ws.ManifestProject.StartBranch("default", "")
		if err != nil {
			return fmt.Errorf("cannot create default in manifest: %s", err)
		}
	}

	err = v.ws.LinkManifest()
	if err != nil {
		return err
	}

	if cap.Isatty() {
		if v.O.ConfigName || v.shouldConfigUser() {
			v.configureUser()
		}
		v.configureColor()
	}

	if v.O.Mirror {
		log.Notef("repo mirror has been initialized in %s", v.ws.RootDir)
	} else {
		log.Notef("repo has been initialized in %s", v.ws.RootDir)
	}

	return nil
}

func (v initCommand) shouldConfigUser() bool {
	var (
		userName  = "user.name"
		userEmail = "user.email"
		err       error
	)

	cfg := v.ws.Config()
	if cfg.Get(userName) == "" || cfg.Get(userEmail) == "" {
		gc, _ := goconfig.GlobalConfig()
		sc, _ := goconfig.SystemConfig()

		if cfg.Get(userName) == "" {
			if gc.Get(userName) != "" {
				cfg.Add(userName, gc.Get(userName))
			} else if sc.Get(userName) != "" {
				cfg.Add(userName, sc.Get(userName))
			} else {
				return true
			}
		}

		if cfg.Get(userEmail) == "" {
			if gc.Get(userEmail) != "" {
				cfg.Add(userEmail, gc.Get(userEmail))
			} else if sc.Get(userEmail) != "" {
				cfg.Add(userEmail, sc.Get(userEmail))
			} else {
				return true
			}
		}

		err = v.ws.SaveConfig(cfg)
		if err != nil {
			log.Error(err)
			return true
		}
	}

	log.Notef("Your identity is: %s <%s>", cfg.Get(userName), cfg.Get(userEmail))
	log.Note("If you want to change this, please re-run 'repo init' with --config-name")

	return false
}

func (v initCommand) userInput(prompt, value string) string {
	if prompt != "" {
		fmt.Printf("%-10s [%s]: ", prompt, value)
	}

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	if text == "" {
		return value
	}
	return text
}

func (v initCommand) configureUser() {
	var (
		userName  = "user.name"
		userEmail = "user.email"
	)

	for {
		cfg := v.ws.Config()
		name := v.userInput("Your Name", cfg.Get(userName))
		email := v.userInput("Your Email", cfg.Get(userEmail))
		fmt.Println("")
		fmt.Printf("Your identity is: %s <%s>", name, email)
		fmt.Printf("is this correct [y/N]? ")
		confirm := strings.ToLower(v.userInput("", "n"))
		if confirm == "y" || confirm == "yes" || confirm == "t" || confirm == "true" || confirm == "on" {
			cfg.Set(userName, name)
			cfg.Set(userEmail, email)
			v.ws.SaveConfig(cfg)
			break
		}
	}
}

func (v initCommand) configureColor() {
	cfg := v.ws.Config()
	for _, k := range []string{"color.ui", "color.diff", "color.status"} {
		if cfg.Get(k) != "" {
			return
		}
	}

	fmt.Println("")
	fmt.Printf("Testing colorized output (for 'repo diff', 'repo status'):\n")

	for _, c := range []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan"} {
		fmt.Print(" ")
		fmt.Printf("%s %-6s %s", color.Color(c, "", ""), c, color.Reset)
	}
	fmt.Print(" ")
	fmt.Printf("%s %s %s", color.Color("white", "black", ""), "white", color.Reset)
	fmt.Println("")

	for _, c := range []string{"bold", "dim", "ul", "reverse"} {
		fmt.Print(" ")
		fmt.Printf("%s %-6s %s", color.Color("black", "", c), c, color.Reset)
	}
	fmt.Println("")

	fmt.Printf("Enable color display in this user account (y/N)? ")
	confirm := strings.ToLower(v.userInput("", "n"))
	if confirm == "y" || confirm == "yes" || confirm == "t" || confirm == "true" || confirm == "on" {
		cfg.Set("color.ui", "auto")
		v.ws.SaveConfig(cfg)
	}
}

var initCmd = initCommand{}

func init() {
	rootCmd.AddCommand(initCmd.Command())
}
