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
	"regexp"
	"sort"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/editor"
	"code.alibaba-inc.com/force/git-repo/project"
	"code.alibaba-inc.com/force/git-repo/workspace"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// UnusualCommitThreshold defines threshold of number of commits to confirm
	UnusualCommitThreshold = 5
)

type uploadCommand struct {
	cmd *cobra.Command
	ws  *workspace.RepoWorkSpace

	O struct {
		AllowAllHooks bool
		AutoTopic     bool
		Branch        string
		BypassHooks   bool
		Cc            []string
		CurrentBranch bool
		Description   string
		DestBranch    string
		Draft         bool
		Issue         string
		NoEmails      bool
		Private       bool
		PushOptions   []string
		Reviewers     []string
		Title         string
		WIP           bool
	}
}

func (v *uploadCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "upload",
		Short: "Upload changes for code review",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runE(args)
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
	v.cmd.Flags().StringVarP(&v.O.DestBranch,
		"dest",
		"D",
		"",
		"Submit for review on this target branch")
	v.cmd.Flags().Bool("no-cert-checks",
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
	v.cmd.Flags().Bool("dryrun",
		false,
		"dryrun mode")

	v.cmd.Flags().Bool("assume-yes",
		false,
		"Automatic yes to prompts")

	v.cmd.Flags().Bool("assume-no",
		false,
		"Automatic no to prompts")

	v.cmd.Flags().Bool("mock-git-push",
		false,
		"Mock git-push for test")

	v.cmd.Flags().String("mock-edit-script",
		"",
		"Mock edit script result file")

	v.cmd.Flags().MarkHidden("auto-topic")
	v.cmd.Flags().MarkHidden("assume-yes")
	v.cmd.Flags().MarkHidden("assume-no")
	v.cmd.Flags().MarkHidden("mock-git-push")
	v.cmd.Flags().MarkHidden("mock-edit-script")

	viper.BindPFlag("no-cert-checks", v.cmd.Flags().Lookup("no-cert-checks"))
	viper.BindPFlag("dryrun", v.cmd.Flags().Lookup("dryrun"))
	viper.BindPFlag("assume-yes", v.cmd.Flags().Lookup("assume-yes"))
	viper.BindPFlag("assume-no", v.cmd.Flags().Lookup("assume-no"))
	viper.BindPFlag("mock-git-push", v.cmd.Flags().Lookup("mock-git-push"))
	viper.BindPFlag("mock-edit-script", v.cmd.Flags().Lookup("mock-edit-script"))

	return v.cmd

}

func (v *uploadCommand) RepoWorkSpace() *workspace.RepoWorkSpace {
	if v.ws == nil {
		v.reloadRepoWorkSpace()
	}
	return v.ws
}

func (v *uploadCommand) reloadRepoWorkSpace() {
	var err error
	v.ws, err = workspace.NewRepoWorkSpace("")
	if err != nil {
		log.Fatal(err)
	}
}

func (v uploadCommand) UploadSingleBranch(branch *project.ReviewableBranch, people [][]string) error {
	var (
		answer bool
	)

	p := branch.Project
	remote := p.Remote.GetRemote()
	key := fmt.Sprintf("review.%s.autoupload", remote.Review)
	commitList := branch.Commits()
	cfg := p.ConfigWithDefault()
	if cfg.HasKey(key) {
		answer = cfg.GetBool(key, false)
		if !answer {
			return fmt.Errorf("upload blocked by %s = false", key)
		}
	} else {
		destBranch := ""
		if v.O.DestBranch != "" {
			destBranch = v.O.DestBranch
		} else if p.DestBranch != "" {
			destBranch = p.DestBranch
		} else if remote.Revision != "" {
			destBranch = remote.Revision
		}

		draftStr := ""
		if v.O.Draft {
			draftStr = " (draft)"
		}
		fmt.Printf("Upload project %s/ to remote branch %s%s:\n",
			p.Path, destBranch, draftStr)
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
		}
		if !answer {
			return fmt.Errorf("upload aborted by user")
		}
	}

	if len(commitList) > UnusualCommitThreshold {
		fmt.Printf("ATTENTION: You are uploading an unusually high number of commits.\n")
		fmt.Println("YOU PROBABLY DO NOT MEAN TO DO THIS. (Did you rebase across branches?)")
		input := userInput("If you are sure you intend to do this, type 'yes': ", "N")
		if answerIsTrue(input) {
			answer = true
		}
		if !answer {
			return fmt.Errorf("upload aborted by user")
		}
	}

	return v.UploadAndReport([]project.ReviewableBranch{*branch}, people)
}

func (v uploadCommand) UploadMultipleBranches(branchesMap map[string][]project.ReviewableBranch, people [][]string) error {
	var (
		projectPattern = regexp.MustCompile(`^#?\s*project\s*([^\s]+)/:$`)
		branchPattern  = regexp.MustCompile(`^\s*branch\s*([^\s(]+)\s*\(.*`)
		ok             bool
	)

	projectsIdx := make(map[string]project.Project)
	branchesIdx := make(map[string]map[string]project.ReviewableBranch)

	keys := []string{}
	for key := range branchesMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	script := []string{"# Uncomment the branches to upload:"}
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
			var destBranch string
			if v.O.DestBranch != "" {
				destBranch = v.O.DestBranch
			} else if branch.Project.DestBranch != "" {
				destBranch = branch.Project.DestBranch
			} else {
				destBranch = branch.Project.Revision
			}
			script = append(script,
				fmt.Sprintf("#  branch %s (%2d commit(s)) to remote branch %s:",
					name,
					len(commitList),
					destBranch))
			for i := range commitList {
				if i < 10 {
					script = append(script,
						fmt.Sprintf("#         %s", commitList[i]))
				} else if i == len(commitList)-1 {
					script = append(script, "#         ... ...")
				}
			}
			b[name] = branch
		}

		projectsIdx[p.Path] = *p
		branchesIdx[p.Name] = b
	}

	editor := editor.Editor{}
	script = append(script, "")
	editString := editor.EditString(strings.Join(script, "\n"))
	if config.MockEditScript() != "" {
		f, err := os.Open(config.MockEditScript())
		if err == nil {
			buf, err := ioutil.ReadAll(f)
			if err == nil {
				editString = string(buf)
			}
		}
	}
	script = strings.Split(editString, "\n")

	todo := []project.ReviewableBranch{}

	var (
		p          project.Project
		hasProject = false
	)
	for _, line := range script {
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

	hasManyCommits := false
	for _, branch := range todo {
		if len(branch.Commits()) > UnusualCommitThreshold {
			hasManyCommits = true
			break
		}
	}
	if hasManyCommits {
		fmt.Println("ATTENTION: One or more branches has an unusually high number of commits.")
		fmt.Println("YOU PROBABLY DO NOT MEAN TO DO THIS. (Did you rebase across branches?)")
		input := userInput("If you are sure you intend to do this, type 'yes': ", "N")
		if !answerIsTrue(input) {
			return fmt.Errorf("upload aborted by user")
		}
	}

	return v.UploadAndReport(todo, people)
}

func (v uploadCommand) UploadAndReport(branches []project.ReviewableBranch, origPeople [][]string) error {

	haveErrors := false
	for _, branch := range branches {
		people := [][]string{[]string{}, []string{}}
		people[0] = append(people[0], origPeople[0]...)
		people[1] = append(people[1], origPeople[1]...)
		branch.AppendReviewers(people)
		isClean, err := branch.Project.IsClean()
		if err != nil {
			log.Error(err)
		}
		cfg := branch.Project.ConfigWithDefault()
		if !isClean {
			key := fmt.Sprintf("review.%s.autoupload", branch.Project.Remote.GetRemote().Review)
			if !cfg.HasKey(key) {
				fmt.Printf("Uncommitted changes in " + branch.Project.Name)
				fmt.Printf(" (did you forget to amend?):\n")
				input := userInput(
					fmt.Sprintf("Continue uploading? (y/N) "),
					"N")
				if answerIsTrue(input) {
					log.Note("skipping upload")
					branch.Uploaded = false
					branch.Error = fmt.Errorf("User aborted")
					continue
				}
			}
		}
		if !v.O.AutoTopic {
			key := fmt.Sprintf("review.%s.uploadtopic", branch.Project.Remote.GetRemote().Review)
			v.O.AutoTopic = cfg.GetBool(key, false)
		}

		destBranch := ""
		if v.O.DestBranch != "" {
			destBranch = v.O.DestBranch
		} else if branch.Project.DestBranch != "" {
			destBranch = branch.Project.DestBranch
		}
		if destBranch != "" {
			fullDest := destBranch
			if !strings.HasPrefix(fullDest, config.RefsHeads) {
				fullDest = config.RefsHeads + fullDest
			}
			mergeBranch := branch.RemoteTrack.Name
			if v.O.DestBranch == "" && mergeBranch != "" && mergeBranch != fullDest {
				fmt.Printf("merge branch %s does not match destination branch %s\n",
					mergeBranch,
					fullDest)
				fmt.Println("skipping upload.")
				fmt.Printf("Please use `--destination %s` if this is intentional\n",
					destBranch)
				branch.Uploaded = false
				continue
			}
		}

		o := project.UploadOptions{
			AutoTopic:    v.O.AutoTopic,
			Description:  v.O.Description,
			DestBranch:   destBranch,
			Draft:        v.O.Draft,
			Issue:        v.O.Issue,
			NoCertChecks: config.NoCertChecks(),
			NoEmails:     v.O.NoEmails,
			Private:      v.O.Private,
			PushOptions:  v.O.PushOptions,
			Title:        v.O.Title,
			WIP:          v.O.WIP,
		}

		err = branch.UploadForReview(&o, people)

		if err != nil {
			branch.Uploaded = false
			branch.Error = err
			haveErrors = true
		} else {
			branch.Uploaded = true
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

func (v uploadCommand) runE(args []string) error {
	ws := v.RepoWorkSpace()
	err := ws.LoadRemotes()
	if err != nil {
		return err
	}

	allProjects, err := ws.GetProjects(nil, args...)
	if err != nil {
		return err
	}

	branch := v.O.Branch

	tasks := make(map[string][]project.ReviewableBranch)
	for _, p := range allProjects {
		if v.O.CurrentBranch {
			cbr := p.GetHead()
			uploadBranch := p.GetUploadableBranch(cbr)
			if uploadBranch != nil {
				tasks[p.Path] = []project.ReviewableBranch{*uploadBranch}
			}
		} else {
			uploadBranches := p.GetUploadableBranches(branch)
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

	people := [][]string{[]string{}, []string{}}
	if len(v.O.Reviewers) > 0 {
		for _, reviewer := range strings.Split(
			strings.Join(v.O.Reviewers, ","),
			",") {
			reviewer = strings.TrimSpace(reviewer)
			if reviewer != "" {
				people[0] = append(people[0], reviewer)
			}
		}
	}
	if len(v.O.Cc) > 0 {
		for _, reviewer := range strings.Split(
			strings.Join(v.O.Cc, ","),
			",") {
			reviewer = strings.TrimSpace(reviewer)
			if reviewer != "" {
				people[1] = append(people[1], reviewer)
			}
		}
	}

	if len(tasks) == 1 {
		for key := range tasks {
			if len(tasks[key]) == 1 {
				return v.UploadSingleBranch(&tasks[key][0], people)
			}
		}
	}
	return v.UploadMultipleBranches(tasks, people)
}

var uploadCmd = uploadCommand{}

func init() {
	rootCmd.AddCommand(uploadCmd.Command())
}
