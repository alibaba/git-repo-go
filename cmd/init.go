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
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/color"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
)

var initOptions = struct {
	ManifestURL       string
	ManifestBranch    string
	CurrentBranchOnly bool
	ManifestName      string
	Mirror            bool
	Reference         string
	Dissociate        bool
	Depth             int
	Archive           bool
	Submodules        bool
	Groups            string
	Platform          string
	NoCloneBundle     bool
	NoTags            bool
	ConfigName        bool
}{}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize manifest repo in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		return initCmdRunE()
	},
}

func initGetGroupStr(ws *workspace.WorkSpace) string {
	allPlatforms := []string{"linux", "darwin", "windows"}
	groups := []string{}
	for _, g := range strings.Split(initOptions.Groups, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			groups = append(groups, g)
		}
	}
	platformize := func(x string) string { return "platform-" + x }
	if initOptions.Platform == "auto" {
		isMirror, _ := ws.ManifestProject.Config().GetBool("repo.mirror", false)
		if !initOptions.Mirror && !isMirror {
			groups = append(groups, platformize(runtime.GOOS))
		}
	} else if initOptions.Platform != "" {
		found := false
		for _, sys := range allPlatforms {
			if initOptions.Platform == "all" || initOptions.Platform == sys {
				groups = append(groups, platformize(sys))
				found = true
			}
		}
		if !found {
			log.Fatalf("invalid platform flag: %s", initOptions.Platform)
		}
	}

	groupStr := strings.Join(groups, ",")
	if initOptions.Platform == "auto" &&
		groupStr == "default,"+platformize(runtime.GOOS) {
		groupStr = ""
	}

	return groupStr
}

func initGuessManifestReference() string {
	var (
		rdir = ""
		err  error
	)

	if initOptions.Reference == "" {
		return ""
	}

	initOptions.Reference, err = path.Abs(initOptions.Reference)
	if err != nil {
		log.Errorf("bad --reference setting: %s", err)
		initOptions.Reference = ""
		return ""
	}

	if initOptions.ManifestURL != "" {
		u, err := url.Parse(initOptions.ManifestURL)
		if err == nil {
			dir := u.RequestURI()
			if !strings.HasSuffix(dir, ".git") {
				dir += ".git"
			}
			dirs := strings.Split(dir, "/")
			for i := 1; i < len(dirs); i++ {
				dir = filepath.Join(initOptions.Reference, filepath.Join(dirs[i:]...))
				if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
					rdir = dir
					break
				}
			}
		}
	}

	if rdir == "" {
		dir := filepath.Join(initOptions.Reference, config.DotRepo, config.ManifestsDotGit)
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			rdir = dir
		}
	}

	return rdir
}

func initCmdRunE() error {
	var (
		err   error
		isNew bool
	)

	if initOptions.Archive && initOptions.Mirror {
		log.Fatal("--mirror and --archive cannot be used together")
	}

	if initOptions.ManifestURL != "" {
		if strings.HasSuffix(initOptions.ManifestURL, "/") {
			initOptions.ManifestURL = strings.TrimRight(initOptions.ManifestURL, "/")
		}
	}

	// Find repo workspace and load it if exist
	ws, err := workspace.NewWorkSpaceInit("", initOptions.ManifestURL)
	if err != nil {
		log.Fatal(err)
	}

	// if workspace is not initialized, check -u, and init it
	if !ws.IsInitialized() {
		isNew = true
		if initOptions.ManifestURL == "" {
			log.Fatal("option --manifest-url (-u) is required")
		}

		ws.ManifestProject.GitInit(initOptions.ManifestURL, initGuessManifestReference())
	} else if initOptions.ManifestURL != "" && initOptions.ManifestURL != ws.ManifestURL() {
		ws.ManifestProject.GitInit(initOptions.ManifestURL, initGuessManifestReference())
	}

	// Fetch repositories
	if initOptions.ManifestBranch != "" {
		ws.ManifestProject.Revision = initOptions.ManifestBranch
	}

	fetchOptions := config.FetchOptions{
		Quiet:             config.GetQuiet(),
		IsNew:             isNew,
		CurrentBranchOnly: initOptions.CurrentBranchOnly,
		CloneBundle:       !initOptions.NoCloneBundle,
		ForceSync:         false,
		NoTags:            initOptions.NoTags,
		Archive:           initOptions.Archive,
		OptimizedFetch:    false,
		Prune:             false,
		Submodules:        initOptions.Submodules,
	}

	err = ws.ManifestProject.Fetch(&fetchOptions)
	if err != nil && isNew &&
		ws.ManifestProject.WorkRepository.Path != "" {
		// Better delete the manifest git dir if we created it; otherwise next
		// time (when user fixes problems) we won't go through the "isNew" logic.
		os.RemoveAll(ws.ManifestProject.WorkRepository.Path)
	}

	// sync repository, and only sync current branch if initOptions.CurrentBranchOnly == true
	// Fetch from remote

	/*
	   if initOptions.manifest_branch:
	     m.MetaBranchSwitch(submodules=initOptions.submodules)

	   syncbuf = SyncBuffer(m.config)
	   m.Sync_LocalHalf(syncbuf, submodules=initOptions.submodules)
	   syncbuf.Finish()

	   if is_new or m.CurrentBranch is None:
	     if not m.StartBranch('default'):
	       print('fatal: cannot create default in manifest', file=sys.stderr)
	       sys.exit(1)

	*/

	// Checkout
	err = ws.ManifestProject.Checkout(initOptions.ManifestBranch, "default")
	if err != nil {
		return err
	}

	err = ws.LinkManifest(initOptions.ManifestName)
	if err != nil {
		return err
	}

	// Save settings to gitcofig of manifest project
	cfg := ws.ManifestProject.Config()
	changed := false

	if initOptions.Reference != "" {
		cfg.Set(config.CfgRepoReference, initOptions.Reference)
		changed = true
	}

	if groupStr := initGetGroupStr(ws); groupStr != cfg.Get(config.CfgManifestGroups) {
		if groupStr != "" {
			cfg.Set(config.CfgManifestGroups, groupStr)
		} else {
			cfg.Unset(config.CfgManifestGroups)
		}
		changed = true
	}

	if initOptions.Dissociate {
		cfg.Set(config.CfgRepoDissociate, true)
		changed = true
	}

	if initOptions.Archive {
		if isNew {
			cfg.Set(config.CfgRepoArchive, true)
			changed = true
		} else {
			log.Fatal(`--archive is only supported when initializing a new workspace.
Either delete the .repo folder in this workspace, or initialize in another location.`)
		}
	}

	if initOptions.Mirror {
		if isNew {
			cfg.Set(config.CfgRepoMirror, true)
			changed = true
		} else {
			log.Fatal(`--mirror is only supported when initializing a new workspace.
Either delete the .repo folder in this workspace, or initialize in another location.`)
		}
	}

	if initOptions.Submodules {
		cfg.Set(config.CfgRepoSubmodules, true)
		changed = true
	}

	if initOptions.ManifestName != "" {
		cfg.Set(config.CfgManifestName, initOptions.ManifestName)
		changed = true
	}

	if initOptions.Depth > 0 {
		cfg.Set(config.CfgRepoDepth, initOptions.Depth)
		changed = true
	}

	if changed {
		err = ws.ManifestProject.SaveConfig(cfg)
		if err != nil {
			return fmt.Errorf("fail to save manifest config: %s", err)
		}
	}

	if cap.Isatty() {
		if initOptions.ConfigName || initShouldConfigUser(ws) {
			initConfigureUser(ws)
		}
		initConfigureColor(ws)
	}

	return nil
}

func initShouldConfigUser(ws *workspace.WorkSpace) bool {
	var (
		userName  = "user.name"
		userEmail = "user.email"
		err       error
	)

	cfg := ws.Config()
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

		err = ws.SaveConfig(cfg)
		if err != nil {
			log.Error(err)
			return true
		}
	}

	log.Notef("Your identity is: %s <%s>", cfg.Get(userName), cfg.Get(userEmail))
	log.Note("If you want to change this, please re-run 'repo init' with --config-name")

	return false
}

func userInput(prompt, value string) string {
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

func initConfigureUser(ws *workspace.WorkSpace) {
	var (
		userName  = "user.name"
		userEmail = "user.email"
	)

	for {
		cfg := ws.Config()
		name := userInput("Your Name", cfg.Get(userName))
		email := userInput("Your Email", cfg.Get(userEmail))
		fmt.Println("")
		fmt.Printf("Your identity is: %s <%s>", name, email)
		fmt.Printf("is this correct [y/N]? ")
		confirm := strings.ToLower(userInput("", "n"))
		if confirm == "y" || confirm == "yes" || confirm == "t" || confirm == "true" || confirm == "on" {
			cfg.Set(userName, name)
			cfg.Set(userEmail, email)
			ws.SaveConfig(cfg)
			break
		}
	}
}

func initConfigureColor(ws *workspace.WorkSpace) {
	cfg := ws.Config()
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
	confirm := strings.ToLower(userInput("", "n"))
	if confirm == "y" || confirm == "yes" || confirm == "t" || confirm == "true" || confirm == "on" {
		cfg.Set("color.ui", "auto")
		ws.SaveConfig(cfg)
	}
}

func init() {
	initCmd.Flags().StringVarP(&initOptions.ManifestURL,
		"manifest-url",
		"u",
		"",
		"manifest repository location")
	initCmd.Flags().StringVarP(&initOptions.ManifestBranch,
		"manifest-branch",
		"b",
		"",
		"manifest branch or revision")
	initCmd.Flags().BoolVar(&initOptions.CurrentBranchOnly,
		"current-branch",
		false, "fetch only current manifest branch from server")
	initCmd.Flags().StringVarP(&initOptions.ManifestName,
		"manifest-name",
		"m",
		"default.xml",
		"initial manifest file")
	initCmd.Flags().BoolVar(&initOptions.Mirror,
		"mirror",
		false,
		"create a replica of the remote repositories rather than a client working directory")
	initCmd.Flags().StringVar(&initOptions.Reference,
		"reference",
		"",
		"location of mirror directory")
	initCmd.Flags().BoolVar(&initOptions.Dissociate,
		"dissociate",
		false,
		"dissociate from reference mirrors after clone")
	initCmd.Flags().IntVar(&initOptions.Depth,
		"depth",
		0,
		"create a shallow clone with given depth; see git clone")
	initCmd.Flags().BoolVar(&initOptions.Archive,
		"archive",
		false,
		"checkout an archive instead of a git repository for each project. See git archive.")
	initCmd.Flags().BoolVar(&initOptions.Submodules,
		"submodules",
		false,
		"sync any submodules associated with the manifest repo")
	initCmd.Flags().StringVarP(&initOptions.Groups,
		"groups",
		"g",
		"",
		"restrict manifest projects to ones with specified group(s) [default|all|G1,G2,G3|G4,-G5,-G6]")
	initCmd.Flags().StringVarP(&initOptions.Platform,
		"platform",
		"p",
		"",
		"restrict manifest projects to ones with a specified platform group [auto|all|none|linux|darwin|...]")
	initCmd.Flags().BoolVar(&initOptions.NoCloneBundle,
		"no-clone-bundle",
		false,
		"disable use of /clone.bundle on HTTP/HTTPS")
	initCmd.Flags().BoolVar(&initOptions.NoTags,
		"no-tags",
		false,
		"don't fetch tags in the manifest")
	initCmd.Flags().BoolVar(&initOptions.ConfigName,
		"config-name",
		false,
		"Always prompt for name/e-mail")

	rootCmd.AddCommand(initCmd)
}
