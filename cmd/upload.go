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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alibaba/git-repo-go/common"
	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/editor"
	"github.com/alibaba/git-repo-go/helper"
	"github.com/alibaba/git-repo-go/path"
	"github.com/alibaba/git-repo-go/project"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// unusualCommitThreshold defines threshold of number of commits to confirm
	unusualCommitThreshold = 5

	// uploadOptionsFile stores upload options to file
	uploadOptionsFile = "UPLOAD_OPTIONS"
	uploadOptionsDir  = "UPLOAD_OPTIONS.d"
)

var (
	reEditSection = regexp.MustCompile(`^#\s+\[(\S+?)\](\s*:.*)?$`)
)

type uploadOptions struct {
	AllowAllHooks  bool
	AutoTopic      bool
	Branch         string
	BypassHooks    bool
	Cc             []string
	CodeReview     common.CodeReview
	CurrentBranch  bool
	Description    string
	DestBranch     string
	Draft          bool
	Issue          string
	MockGitPush    bool
	MockEditScript string
	NoCache        bool
	NoCertChecks   bool
	NoEdit         bool
	NoEmails       bool
	Private        bool
	PushOptions    []string
	Reviewers      []string
	Remote         string
	Title          string
	WIP            bool
}

// LoadFromFile reads content from file and parses into push options.
func (v *uploadOptions) LoadFromFile(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	v.LoadFromText(string(data))
}

// LoadFromText parses content into upload options.
func (v *uploadOptions) LoadFromText(data string) {
	var (
		section string
		text    string
	)

	setUploadOption := func(section, text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		switch section {
		case "title":
			text = strings.Split(text, "\n")[0]
			v.Title = text
		case "description":
			v.Description = text
		case "issue":
			v.Issue = strings.Join(strings.Split(text, "\n"), ",")
		case "reviewer":
			v.Reviewers = strings.Split(text, "\n")
		case "cc":
			v.Cc = strings.Split(text, "\n")
		case "draft", "private":
			switch text {
			case "y", "yes", "on", "t", "true", "1":
				if section == "draft" {
					v.Draft = true
				} else if section == "private" {
					v.Private = true
				}
			case "n", "no", "off", "f", "false", "0":
				if section == "draft" {
					v.Draft = false
				} else if section == "private" {
					v.Private = false
				}
			default:
				log.Warnf("cannot turn '%s' to boolean", text)
			}
		default:
			log.Warnf("unknown section name: %s", section)
		}
	}

	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimRight(line, " \t")
		if m := reEditSection.FindStringSubmatch(line); m != nil {
			name := strings.ToLower(m[1])
			switch name {
			case
				"title",
				"description",
				"issue",
				"reviewer",
				"cc",
				"draft",
				"private":

				if section != "" {
					setUploadOption(section, text)
				}
				section = name
				text = ""
				continue
			default:
				log.Warnf("unknown section '%s' in script", name)
			}
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if section != "" {
			text += line + "\n"
		}
	}

	if section != "" {
		setUploadOption(section, text)
	}
}

// Export will export uploadOptions for edit.
func (v *uploadOptions) Export(published bool) []string {
	script := []string{}
	w := 13
	if !published {
		script = append(script, fmt.Sprintf("# %-*s : %s", w,
			"[Title]",
			"one line message below as the title of code review"),
		)
		if v.Title != "" {
			script = append(script, "", v.Title)
		}
		script = append(script, "")

		script = append(script, fmt.Sprintf("# %-*s : %s", w,
			"[Description]",
			"multiple lines of text as the description of code review"),
		)
		if v.Description != "" {
			script = append(script, "")
			script = append(script, strings.Split(v.Description, "\n")...)
		}
		script = append(script, "")
	}

	script = append(script, fmt.Sprintf("# %-*s : %s", w,
		"[Issue]",
		"multiple lines of issue IDs for cross references"),
	)
	if v.Issue != "" {
		script = append(script, "", v.Issue)
	}
	script = append(script, "")

	script = append(script, fmt.Sprintf("# %-*s : %s", w,
		"[Reviewer]",
		"multiple lines of user names as the reviewers for code review"),
	)
	if len(v.Reviewers) > 0 {
		script = append(script, "")
		script = append(script, v.Reviewers...)
	}
	script = append(script, "")

	script = append(script, fmt.Sprintf("# %-*s : %s", w,
		"[Cc]",
		"multiple lines of user names as the watchers for code review"),
	)
	if len(v.Cc) > 0 {
		script = append(script, "")
		script = append(script, v.Cc...)
	}
	script = append(script, "")

	script = append(script, fmt.Sprintf("# %-*s : %s", w,
		"[Draft]",
		"a boolean (yes/no, or true/false) to turn on/off draft mode"),
	)
	if v.Draft {
		script = append(script, "", "yes")
	}
	script = append(script, "")

	script = append(script, fmt.Sprintf("# %-*s : %s", w,
		"[Private]",
		"a boolean (yes/no, or true/false) to turn on/off private mode"),
	)
	if v.Private {
		script = append(script, "", "yes")
	}
	script = append(script, "")

	return script
}

type uploadCommand struct {
	WorkSpaceCommand

	cmd *cobra.Command
	O   uploadOptions
}

func (v *uploadCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload changes for code review",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	aliasNormalizeFunc := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "re":
			name = "reviewers"
		case "reviewer":
			name = "reviewers"
		case "current-branch":
			name = "cbr"
		case "destination":
			name = "dest"
		}
		return pflag.NormalizedName(name)
	}

	v.cmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)
	v.cmd.Flags().BoolVarP(&v.O.AutoTopic,
		"auto-topic",
		"t",
		false,
		"Send local branch name for Code Review")
	v.cmd.Flags().StringArrayVar(&v.O.Reviewers,
		"reviewers",
		nil,
		"Request reviews from these people")
	v.cmd.Flags().StringArrayVar(&v.O.Cc,
		"cc",
		nil,
		"Also send email to these email addresses")
	v.cmd.Flags().StringVar(&v.O.Branch,
		"br",
		"",
		"Branch to upload")
	v.cmd.Flags().StringVarP(&v.O.CodeReview.ID,
		"change",
		"c",
		"",
		"ID of the specific code review to change")
	v.cmd.Flags().BoolVar(&v.O.CurrentBranch,
		"cbr",
		false,
		"Upload current git branch")
	v.cmd.Flags().BoolVarP(&v.O.Draft,
		"draft",
		"d",
		false,
		"If specified, upload as a draft")
	v.cmd.Flags().BoolVar(&v.O.NoEmails,
		"no-emails",
		false,
		"If specified, do not send emails on upload")
	v.cmd.Flags().BoolVarP(&v.O.Private,
		"private",
		"p",
		false,
		"If specified, upload as a private change")
	v.cmd.Flags().StringVar(&v.O.Title,
		"title",
		"",
		"Title for review")
	v.cmd.Flags().StringVar(&v.O.Description,
		"description",
		"",
		"Description for review")
	v.cmd.Flags().StringVar(&v.O.Issue,
		"issue",
		"",
		"Related issues for review")
	v.cmd.Flags().BoolVarP(&v.O.WIP,
		"wip",
		"w",
		false,
		"If specified, upload as a work-in-progress change")
	v.cmd.Flags().StringArrayVarP(&v.O.PushOptions,
		"push-options",
		"o",
		nil,
		"Additional push options to transmit")
	v.cmd.Flags().StringVar(&v.O.Remote,
		"remote",
		"",
		"use specific remote for upload (use with --single)")
	v.cmd.Flags().StringVarP(&v.O.DestBranch,
		"dest",
		"D",
		"",
		"Submit for review on this target branch")
	v.cmd.Flags().BoolVar(&v.O.NoCertChecks,
		"no-cert-checks",
		false,
		"Disable verifying ssl certs (unsafe)")
	v.cmd.Flags().BoolVar(&v.O.BypassHooks,
		"no-verify",
		false,
		"Do not run the upload hook")
	v.cmd.Flags().BoolVar(&v.O.AllowAllHooks,
		"verify",
		false,
		"Run the upload hook without prompting")
	v.cmd.Flags().BoolVar(&v.O.NoCache,
		"no-cache",
		false,
		"Ignore ssh-info cache, and recheck ssh-info API")

	v.cmd.Flags().BoolVar(&v.O.NoEdit,
		"no-edit",
		false,
		"If specified, do not open editor to confirm")
	v.cmd.Flags().BoolVar(&v.O.MockGitPush,
		"mock-git-push",
		false,
		"Mock git-push for test")
	v.cmd.Flags().StringVar(&v.O.MockEditScript,
		"mock-edit-script",
		"",
		"Mock edit script result file")

	v.cmd.Flags().MarkHidden("auto-topic")
	v.cmd.Flags().MarkHidden("mock-git-push")
	v.cmd.Flags().MarkHidden("mock-edit-script")

	return v.cmd
}

func (v *uploadCommand) getDestBranch(branch *project.ReviewableBranch) (string, error) {
	var (
		destBranch string
	)

	if branch == nil {
		return "", fmt.Errorf("reviewable branch is nil")
	}
	p := branch.Project
	if p == nil {
		return "", fmt.Errorf("project of reviewable branch is nil")
	}
	if v.O.DestBranch != "" {
		destBranch = v.O.DestBranch
	} else if p.DestBranch != "" {
		destBranch = p.DestBranch
	} else if p.Revision != "" {
		destBranch = p.Revision
	}
	return destBranch, nil
}

func (v uploadCommand) UploadForReviewWithConfirm(branchesMap map[string][]project.ReviewableBranch) error {
	var (
		answer   bool
		count    int
		i        int
		branches []project.ReviewableBranch
		todo     []project.ReviewableBranch
	)

	for key := range branchesMap {
		for _, branch := range branchesMap[key] {
			branches = append(branches, branch)
		}
	}
	sort.Slice(branches, func(i, j int) bool {
		if branches[i].Project.Name < branches[j].Project.Name {
			return true
		} else if branches[i].Project.Name == branches[j].Project.Name {
			return branches[i].Branch.Name < branches[j].Branch.Name
		}
		return false
	})
	count = len(branches)

	for _, branch := range branches {
		p := branch.Project
		remote := branch.Remote
		if count > 1 {
			i++
			fmt.Printf("[%d/%d] project %s: %s\n", i, count, p.Path, branch.Branch.Name)
		}
		if remote == nil {
			log.Errorf("cannot find remote of branch '%s' of project '%s'",
				branch.Branch.Name,
				p.Name,
			)
			continue
		}
		commitList := branch.Commits()
		cfg := p.ConfigWithDefault()
		key := fmt.Sprintf("review.%s.autoupload", remote.Review)
		if cfg.HasKey(key) {
			answer = cfg.GetBool(key, false)
			if !answer {
				log.Errorf("upload blocked by %s = false", key)
				continue
			}
		} else {
			draftStr := ""
			if v.O.Draft {
				draftStr = " (draft)"
			}

			if branch.CodeReview.Empty() {
				destBranch, err := v.getDestBranch(&branch)
				if err != nil {
					return err
				}
				if p.Path == "." {
					fmt.Printf("Upload project (%s) to remote branch %s%s:\n",
						p.Name, destBranch, draftStr)
				} else {
					fmt.Printf("Upload project %s/ to remote branch %s%s:\n",
						p.Path, destBranch, draftStr)
				}
			} else {
				fmt.Printf("Upload code review #%s of project (%s)%s:\n",
					branch.CodeReview.ID, p.Name, draftStr)
			}
			fmt.Printf("  branch %s (%2d commit(s)):\n",
				branch.Branch.Name,
				len(commitList))
			for _, commit := range commitList {
				fmt.Printf("         %s\n", commit)
			}

			input := userInput(
				fmt.Sprintf("to %s (y/N)? ", remote.Review),
				"N")
			if answerIsTrue(input) {
				answer = true
				todo = append(todo, branch)
			} else {
				log.Error("upload aborted by user")
				continue
			}
		}

		if len(commitList) > unusualCommitThreshold {
			fmt.Printf("ATTENTION: You are uploading an unusually high number of commits.\n")
			fmt.Println("YOU PROBABLY DO NOT MEAN TO DO THIS. (Did you rebase across branches?)")
			input := userInput("If you are sure you intend to do this, type 'yes': ", "N")
			if !answerIsTrue(input) {
				log.Error("upload aborted by user")
				continue
			}
		}
	}

	if len(todo) > 0 {
		return v.UploadAndReport(todo)
	}
	return fmt.Errorf("nothing confirmed for upload")
}

func (v uploadCommand) UploadForReviewWithEditor(branchesMap map[string][]project.ReviewableBranch) error {
	var (
		projectPattern = regexp.MustCompile(`^#?\s*project\s*([^\s]+)/:$`)
		branchPattern  = regexp.MustCompile(`^\s*branch\s*([^\s(]+)\s*\(.*`)
		ok             bool
		err            error
		branchComment  string
	)

	projectsIdx := make(map[string]project.Project)
	branchesIdx := make(map[string]map[string]project.ReviewableBranch)

	keys := []string{}
	for key := range branchesMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	if config.AssumeYes() {
		branchComment = " "
	} else if config.AssumeNo() {
		branchComment = "#"
	} else if len(keys) == 1 && len(branchesMap[keys[0]]) == 1 {
		branchComment = " "
	} else {
		branchComment = "#"
	}

	// Script for branches selection
	markbranchSelection := "# Step 2: Select project and branches for upload"
	script := []string{
		"",
		"##############################################################################",
		markbranchSelection,
		"#",
		"# Note: Uncomment the branches to upload, and not touch the project lines",
		"##############################################################################",
		"",
	}
	published := true
	optionsFile := ""
	for _, key := range keys {
		branches := branchesMap[key]
		p := branches[0].Project
		script = append(script, "#")
		script = append(script, fmt.Sprintf("# project %s/:", p.Path))

		b := make(map[string]project.ReviewableBranch)
		for _, branch := range branches {
			name := branch.Branch.Name
			// date := branch.date
			commitList := branch.Commits()

			if len(b) > 0 {
				script = append(script, "#")
			}

			if branch.CodeReview.Empty() {
				destBranch, err := v.getDestBranch(&branch)
				if optionsFile == "" {
					optionsFile = destBranch
				}
				if err != nil {
					return err
				}
				script = append(script,
					fmt.Sprintf("%s  branch %s (%2d commit(s)) to remote branch %s:",
						branchComment,
						name,
						len(commitList),
						destBranch))
			} else {
				script = append(script,
					fmt.Sprintf("%s  branch %s (%2d commit(s)) to update code review #%s:",
						branchComment,
						name,
						len(commitList),
						branch.CodeReview.ID))

			}
			for i := range commitList {
				if i < 10 {
					script = append(script,
						fmt.Sprintf("#         %s", commitList[i]))
				} else if i == len(commitList)-1 {
					script = append(script, "#         ... ...")
				}
			}
			if !branch.IsPublished() {
				published = false
			}
			b[name] = branch
		}

		projectsIdx[p.Path] = *p
		branchesIdx[p.Name] = b
	}
	script = append(script, "")

	if strings.HasPrefix(optionsFile, config.RefsHeads) {
		optionsFile = strings.TrimPrefix(optionsFile, config.RefsHeads)
	}
	optionsFile = strings.Replace(optionsFile, "/", ".", -1)
	optionsFile = filepath.Join(v.ws.AdminDir(), uploadOptionsDir, optionsFile)

	script = append(v.fmtUploadOptionsScript(optionsFile, published), script...)

	editString := editor.EditString(strings.Join(script, "\n"))

	if v.O.MockEditScript != "" {
		f, err := os.Open(v.O.MockEditScript)
		if err == nil {
			buf, err := ioutil.ReadAll(f)
			if err == nil {
				editString = string(buf)
			}
		}
	}

	// Load upload options
	optsInEditString := strings.Split(editString, markbranchSelection)[0]
	v.O.LoadFromText(optsInEditString)

	// Save editString to template file `.git/UPLOAD_OPTIONS.d/<branch>`
	err = v.saveUploadOptions(optionsFile, v.O)
	if err != nil {
		log.Error(err)
	}

	// Parse script for branches selection
	script = strings.Split(editString, "\n")
	todo := []project.ReviewableBranch{}

	var (
		p                 project.Project
		hasProject        = false
		inBranchSelection = false
	)
	for _, line := range script {
		if !inBranchSelection {
			if line == markbranchSelection {
				inBranchSelection = true
			}
			continue
		}

		if m := projectPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if p, ok = projectsIdx[name]; !ok {
				log.Fatalf("project %s not available for upload", name)
			}
			hasProject = true
			continue
		}

		if m := branchPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !hasProject {
				log.Fatalf("project for branch %s not in script", name)
			}
			if branch, ok := branchesIdx[p.Name][name]; ok {
				todo = append(todo, branch)
			} else {
				log.Fatalf("branch %s not in %s", name, p.Path)
			}
		}
	}
	if len(todo) == 0 {
		log.Fatal("nothing uncommented for upload")
	}

	return v.UploadAndReport(todo)
}

func (v uploadCommand) saveUploadOptions(optionsFile string, o uploadOptions) error {
	var (
		oldOptions uploadOptions
	)

	if path.Exist(optionsFile) {
		oldOptions.LoadFromFile(optionsFile)
		if oldOptions.Title != "" {
			o.Title = oldOptions.Title
		}
		if oldOptions.Description != "" {
			o.Description = oldOptions.Description
		}
	} else {
		dir := filepath.Dir(optionsFile)
		if !path.Exist(dir) {
			os.Mkdir(dir, 0755)
		}
	}

	lockFile := optionsFile + ".lock"
	data := strings.Join(o.Export(false), "\n")
	err := ioutil.WriteFile(lockFile, []byte(data), 0644)
	if err != nil {
		return err
	}
	return os.Rename(lockFile, optionsFile)
}

func (v uploadCommand) fmtUploadOptionsScript(optionsFile string, published bool) []string {
	var (
		o      = uploadOptions{}
		script = []string{}
	)

	if published {
		script = append(script,
			"##############################################################################",
			"# Step 1: Input your options for code review",
			"#",
			"# Note: Input your options below the comments and keep the comments unchanged,",
			"#       and options which work only for new created code review are hidden.",
			"##############################################################################",
			"",
		)
	} else {
		script = append(script,
			"##############################################################################",
			"# Step 1: Input your options for code review",
			"#",
			"# Note: Input your options below the comments and keep the comments unchanged",
			"##############################################################################",
			"",
		)
	}

	// Load upload options file created by last upload
	if !path.Exist(optionsFile) {
		optionsFile = filepath.Join(v.ws.AdminDir(), uploadOptionsFile)
		if !path.Exist(optionsFile) {
			// fist file in uploadOptionsDir
			filepath.Walk(filepath.Join(v.ws.AdminDir(),
				uploadOptionsDir),
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info != nil && info.IsDir() {
						if filepath.Base(path) == uploadOptionsDir {
							return nil
						}
						return filepath.SkipDir
					}
					// return the first file in uploadOptionsDir
					optionsFile = path
					return filepath.SkipDir
				},
			)
		}

		// fallback to ~/.git-repo/UPLOAD_OPTIONS
		if !path.Exist(optionsFile) {
			configDir, err := config.GetConfigDir()
			if err != nil {
				log.Warnf("fail get config dir: %s", err)
			} else {
				optionsFile = filepath.Join(configDir, uploadOptionsFile)
			}
		}
	}

	buf, err := ioutil.ReadFile(optionsFile)
	if err == nil {
		o.LoadFromText(string(buf))

		if v.O.Title == "" {
			v.O.Title = o.Title
		}
		if v.O.Description == "" {
			v.O.Description = o.Description
		}
		if v.O.Issue == "" {
			v.O.Issue = o.Issue
		}
		if len(v.O.Reviewers) == 0 {
			v.O.Reviewers = o.Reviewers
		}
		if len(v.O.Cc) == 0 {
			v.O.Cc = o.Cc
		}
		if !v.O.Draft {
			v.O.Draft = o.Draft
		}
		if !v.O.WIP {
			v.O.WIP = o.WIP
		}
		if !v.O.Private {
			v.O.Private = o.Private
		}
	}

	script = append(script, v.O.Export(published)...)
	return script
}

func (v *uploadCommand) UploadAndReport(branches []project.ReviewableBranch) error {
	var (
		origPeople = [][]string{[]string{}, []string{}}
		oldOid     = ""
		err        error
		destBranch string
	)

	if len(v.O.Reviewers) > 0 {
		for _, reviewer := range strings.Split(
			strings.Join(v.O.Reviewers, ","),
			",") {
			reviewer = strings.TrimSpace(reviewer)
			if reviewer != "" {
				origPeople[0] = append(origPeople[0], reviewer)
			}
		}
	}
	if len(v.O.Cc) > 0 {
		for _, reviewer := range strings.Split(
			strings.Join(v.O.Cc, ","),
			",") {
			reviewer = strings.TrimSpace(reviewer)
			if reviewer != "" {
				origPeople[1] = append(origPeople[1], reviewer)
			}
		}
	}

	haveErrors := false
	for i := range branches {
		// Will update branch.Error in this loop.
		branch := &(branches[i])
		theProject := branch.Project
		remote := branch.Remote
		if remote == nil {
			log.Errorf("cannot get remote of branch '%s' of project '%s'",
				branch.Branch.Name,
				theProject.Name,
			)
			continue
		}
		people := [][]string{[]string{}, []string{}}
		people[0] = append(people[0], origPeople[0]...)
		people[1] = append(people[1], origPeople[1]...)
		branch.AppendReviewers(people)
		cfg := theProject.ConfigWithDefault()
		if !theProject.IsClean() {
			key := fmt.Sprintf("review.%s.autoupload", remote.Review)
			if !cfg.HasKey(key) {
				fmt.Printf("Uncommitted changes in " + theProject.Name)
				fmt.Printf(" (did you forget to amend?):\n")
				input := userInput(
					fmt.Sprintf("Continue uploading? (y/N) "),
					"N")
				if !answerIsTrue(input) {
					log.Note("skipping upload")
					branch.Uploaded = false
					branch.Error = fmt.Errorf("User aborted")
					continue
				}
			}
		}
		if !v.O.AutoTopic {
			key := fmt.Sprintf("review.%s.uploadtopic", remote.Review)
			v.O.AutoTopic = cfg.GetBool(key, false)
		}

		if v.O.CodeReview.Empty() {
			oldOid = theProject.PublishedRevision(branch.Branch.Name)

			destBranch, err = v.getDestBranch(branch)
			if err != nil {
				return err
			}
			if destBranch != "" {
				fullDest := destBranch
				if !strings.HasPrefix(fullDest, config.RefsHeads) {
					fullDest = config.RefsHeads + fullDest
				}
				mergeBranch := branch.RemoteTrack.Branch
				if !strings.HasPrefix(mergeBranch, config.RefsHeads) {
					mergeBranch = config.RefsHeads + mergeBranch
				}
				if v.O.DestBranch == "" && mergeBranch != "" && mergeBranch != fullDest {
					log.Errorf("merge branch %s does not match destination branch %s\n",
						mergeBranch,
						fullDest)
					log.Errorf("skipping upload.")
					log.Errorf("Please use `--dest %s` if this is intentional\n",
						destBranch)
					branch.Uploaded = false
					continue
				}
			}
		} else {
			oldOid, err = theProject.ResolveRevision(v.O.CodeReview.Ref)
			if err != nil {
				return fmt.Errorf("fail to parse ref '%s', not downloaded yet?",
					v.O.CodeReview.Ref,
				)
			}
		}

		o := common.UploadOptions{
			AutoTopic:    v.O.AutoTopic,
			CodeReview:   v.O.CodeReview,
			Description:  v.O.Description,
			DestBranch:   destBranch,
			Draft:        v.O.Draft,
			Issue:        v.O.Issue,
			LocalBranch:  branch.Branch.Name,
			MockGitPush:  v.O.MockGitPush,
			NoCertChecks: v.O.NoCertChecks || config.NoCertChecks(),
			NoEmails:     v.O.NoEmails,
			OldOid:       oldOid,
			People:       people,
			Private:      v.O.Private,
			PushOptions:  v.O.PushOptions,
			Title:        v.O.Title,
			WIP:          v.O.WIP,
		}

		err = branch.UploadForReview(&o)
		if err != nil {
			branch.Uploaded = false
			branch.Error = err
			haveErrors = true
		} else {
			branch.Uploaded = true
			// Disable default push for single repo workspace,
			// because for multple repository, push.default has
			// already been disabled in `git repo sync` process.
			if v.ws.IsSingle() && theProject != nil {
				// push command must have specific refspec
				theProject.DisableDefaultPush()
			}
		}
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "----------------------------------------------------------------------")
	if haveErrors {
		for _, branch := range branches {
			if !branch.Uploaded && branch.Error != nil {
				format := ""
				if len(branch.Error.Error()) <= 30 {
					format = " (%s)"
				} else {
					format = "\n       (%s)"
				}
				fmt.Fprintf(os.Stderr,
					"[FAILED] %-15s %-15s"+format+"\n",
					branch.Project.Path+"/",
					branch.Branch.Name,
					branch.Error.Error())
			}
		}
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}
	return nil
}

func (v uploadCommand) Execute(args []string) error {
	ws := v.WorkSpace()
	err := ws.LoadRemotes(v.O.NoCache)
	if err != nil {
		return err
	}

	if v.O.Remote != "" && !config.IsSingleMode() {
		return fmt.Errorf("--remote can be only used with --single")
	}

	allProjects, err := ws.GetProjects(nil, args...)
	if err != nil {
		return err
	}

	if len(allProjects) == 0 {
		log.Note("no projects ready for upload")
		return nil
	}

	tasks := make(map[string][]project.ReviewableBranch)
	for _, p := range allProjects {
		// For single project.
		if config.IsSingleMode() {
			var (
				uploadBranch   *project.ReviewableBranch
				head           string
				remoteName     string
				remoteRevision string
				remoteURL      string
				remote         *project.Remote
			)

			if v.O.Branch == "" {
				v.O.CurrentBranch = true
				head = p.GetHead()
			} else {
				head = v.O.Branch
				if !strings.HasPrefix(head, config.RefsHeads) {
					head = config.RefsHeads + head
				}
			}
			if !project.IsHead(head) {
				log.Debugf("detached at %s", head)
				return fmt.Errorf("upload failed: not in a branch\n\n" +
					"Please run command \"git checkout -b <branch>\" to create a new branch.")
			}
			head = strings.TrimPrefix(head, config.RefsHeads)

			if v.O.Remote == "" {
				remote = p.GetBranchRemote(head, true)
				if remote == nil {
					return fmt.Errorf("no remote for branch '%s' of project '%s' to push",
						head,
						p.Name,
					)
				}
				remoteName = remote.Name
			} else {
				remoteName = v.O.Remote
				remote = p.Remotes.Get(remoteName)
				if remote == nil {
					return fmt.Errorf("cannot file remote named '%s'for project '%s'",
						remoteName,
						p.Name,
					)
				}
			}
			if !remote.ProtoHelperReady() {
				return fmt.Errorf("remote '%s' for project '%s' is not reviewable",
					remote.Name,
					p.Name,
				)
			}

			if v.O.CodeReview.ID != "" {
				v.O.CodeReview.Ref, err = remote.GetDownloadRef(v.O.CodeReview.ID, "")
				if err != nil {
					return fmt.Errorf("fail to get local ref for code review #%s: %s",
						v.O.CodeReview.ID,
						err)
				}
			}

			if v.O.DestBranch == "" {
				remoteRevision = p.TrackBranch(head)
			} else {
				remoteRevision = v.O.DestBranch
			}
			if remoteRevision == "" && v.O.CodeReview.Empty() {
				return fmt.Errorf("upload failed: cannot find tracking branch\n\n" +
					"Please run command \"git branch -u <upstream>\" to track a remote branch. E.g.:\n\n" +
					"    git branch -u origin/master\n\n" +
					"Or give the following options when uploading:\n\n" +
					"    --dest <dest-branch> [--remote <remote>]")
			}

			// Set Revision of manifest.Remote to tracking branch.
			// p.Remote.Revision = remoteRevision

			// Set project and repository name
			remoteURL = p.GitConfigRemoteURL(remoteName)
			gitURL := config.ParseGitURL(remoteURL)
			if gitURL != nil && gitURL.Repo != "" {
				if gitURL.Proto == "file" {
					p.Name = filepath.Base(gitURL.Repo)
				} else {
					p.Name = gitURL.Repo
				}
			}

			// Set other missing fields
			p.RemoteURL = remoteURL
			p.RemoteName = remoteName
			p.Revision = remoteRevision

			// Install hooks if remote is Gerrit server
			if remote.GetType() == helper.ProtoTypeGerrit {
				allProjects[0].InstallGerritHooks()
			}

			/////////////
			if v.O.CodeReview.Empty() {
				uploadBranch = p.GetUploadableBranch(head, remote, remoteRevision)
			} else {
				uploadBranch = p.GetUploadableBranchForChange(head, remote, &v.O.CodeReview)
			}
			if uploadBranch != nil {
				tasks[p.Path] = []project.ReviewableBranch{*uploadBranch}
			}
			// No other projects
			break
		}

		// For projects managed by manifests project.
		if v.O.CurrentBranch {
			cbr := p.GetHead()
			remote := p.GetBranchRemote(cbr, false)
			if cbr != "" && remote != nil {
				uploadBranch := p.GetUploadableBranch(cbr, remote, "")
				if uploadBranch != nil {
					tasks[p.Path] = []project.ReviewableBranch{*uploadBranch}
				}
			}
		} else {
			uploadBranches := p.GetUploadableBranches(v.O.Branch)
			if len(uploadBranches) == 0 {
				continue
			}
			tasks[p.Path] = uploadBranches
		}
	}

	if len(tasks) == 0 {
		log.Note("no branches ready for upload")
		return nil
	}

	if v.O.NoEdit || editor.Editor() == "" {
		err = v.UploadForReviewWithConfirm(tasks)
	} else {
		err = v.UploadForReviewWithEditor(tasks)
	}
	if err != nil {
		return err
	}

	// For single mode, clean published refs, because we don't have chance to
	// run other commands, such as `git-repo sync`.
	if config.IsSingleMode() {
		err = allProjects[0].CleanPublishedCache()
	}
	return err
}

var uploadCmd = uploadCommand{
	WorkSpaceCommand: WorkSpaceCommand{
		MirrorOK: false,
		SingleOK: true,
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd.Command())
}
