package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"code.alibaba-inc.com/force/git-repo/path"
	"github.com/jiangxin/goconfig"
	"github.com/jiangxin/multi-log"
)

type keywordSubstFilterDriver struct {
	Filename   string
	Keywords   []string
	Re         *regexp.Regexp
	KeywordMap map[string]string
}

func newKeywordSubstFilterDriver(filename string) filterDriver {
	kf := keywordSubstFilterDriver{
		Filename: filename,
		Keywords: []string{
			"Date",
			"LastChangedDate",
			"Revision",
			"LastChangedRevision",
			"Author",
			"LastChangedBy",
			"HeadURL",
			"Id",
			"Header",
		},
	}

	pattern := "[$](" + strings.Join(kf.Keywords, "|") + ")(:[^$]*)?[$]"
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Errorf("fail to compile regex pattern: %s", err)
	} else {
		kf.Re = re
	}

	return &kf
}

func (v *keywordSubstFilterDriver) Clean() error {
	var (
		err error
	)

	// Bad regexp
	if v.Re == nil {
		_, err = io.Copy(os.Stdout, os.Stdin)
		return err
	}

	r := bufio.NewReader(os.Stdin)
	for {
		buf, err := r.ReadBytes('\n')
		if len(buf) > 0 {
			matches := v.Re.FindAllSubmatch(buf, -1)
			if len(matches) > 0 {
				buf = []byte(v.Re.ReplaceAllString(string(buf), `$$$1$$`))
			}
			_, err = os.Stdout.Write(buf)
			if err != nil {
				log.Errorf("fail to write stdout: %s", err)
			}
		}

		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("unknown error: %s", err)
			break
		}
	}

	return nil
}

func (v *keywordSubstFilterDriver) Smudge() error {
	var (
		err error
	)

	// Bad regexp
	if v.Re == nil {
		_, err = io.Copy(os.Stdout, os.Stdin)
		return err
	}

	r := bufio.NewReader(os.Stdin)
	for {
		buf, err := r.ReadBytes('\n')
		if len(buf) > 0 {
			matches := v.Re.FindAllSubmatch(buf, -1)
			for _, match := range matches {
				buf = v.replace(buf, match)
			}

			_, err = os.Stdout.Write(buf)
			if err != nil {
				log.Errorf("fail to write stdout: %s", err)
			}
		}

		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("unknown error: %s", err)
			break
		}
	}

	return nil
}

func (v *keywordSubstFilterDriver) initialKeywordMap() {
	v.KeywordMap = make(map[string]string)

	cmdArgs := []string{
		"git",
		"log",
		"-1",
		"--no-color",
		"--no-decorate",
		"--pretty=fuller",
		"--",
		v.Filename,
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = nil
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Debugf("fail to get log for filter, maybe current branch is unborn: %s", string(exitError.Stderr))
		}

		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "commit ") {
			v.KeywordMap["Commit"] = strings.TrimPrefix(line, "commit ")
		} else if strings.HasPrefix(line, "Author:") {
			v.KeywordMap["Author"] = strings.TrimSpace(strings.TrimPrefix(line, "Author:"))
			v.KeywordMap["LastChangedBy"] = v.KeywordMap["Author"]
		} else if strings.HasPrefix(line, "CommitDate:") {
			commitDate := strings.TrimSpace(strings.TrimPrefix(line, "CommitDate:"))
			t, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", commitDate)
			if err != nil {
				log.Warnf("fail to parse date: %s", commitDate)
				v.KeywordMap["UTC"] = commitDate
				v.KeywordMap["Date"] = commitDate
				v.KeywordMap["LastChangedDate"] = v.KeywordMap["Date"]
			} else {
				v.KeywordMap["UTC"] = t.UTC().Format("2006-01-02 15:04:05 -0700")
				v.KeywordMap["Date"] = t.Local().Format("2006-01-02 15:04:05 -0700")
				v.KeywordMap["LastChangedDate"] = v.KeywordMap["Date"]
			}
		}
	}

	cmdArgs = []string{
		"git",
		"describe",
		"--always",
		v.KeywordMap["Commit"],
		"--",
	}

	cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = nil
	out, err = cmd.Output()
	if err == nil {
		v.KeywordMap["Revision"] = string(bytes.TrimSpace(out))
		v.KeywordMap["LastChangedRevision"] = v.KeywordMap["Revision"]
	} else {
		log.Debugf("fail to run '%s' to get LastChangedRevision: %s", strings.Join(cmdArgs, " "), err)
	}

	v.KeywordMap["Id"] = fmt.Sprintf("%s %s %s %s",
		filepath.Base(v.Filename),
		v.KeywordMap["Revision"],
		v.KeywordMap["UTC"],
		v.KeywordMap["Author"],
	)

	fullPath := v.Filename
	worktree, _, err := path.FindGitWorkSpace("")
	if err != nil {
		abs, _ := path.Abs(v.Filename)
		fullPath, _ = filepath.Rel(worktree, abs)
	}
	cfg, err := goconfig.Load("")
	if err == nil {
		for _, section := range cfg.Sections() {
			if strings.HasPrefix(section, "remote.") {
				serverURL := cfg.Get(section + ".url")
				fullPath = filepath.Join(serverURL, fullPath)
				break
			}
		}
	}

	v.KeywordMap["Header"] = fmt.Sprintf("%s %s %s %s",
		fullPath,
		v.KeywordMap["Revision"],
		v.KeywordMap["UTC"],
		v.KeywordMap["Author"],
	)

	v.KeywordMap["HeadURL"] = fullPath
}

func (v *keywordSubstFilterDriver) replace(buf []byte, match [][]byte) []byte {
	if v.KeywordMap == nil {
		v.initialKeywordMap()
	}

	keyword := match[1]
	value := v.KeywordMap[string(keyword)]
	if value != "" {
		replace := "$" + string(keyword) + ": " + v.KeywordMap[string(keyword)] + " $"
		buf = bytes.Replace(buf, match[0], []byte(replace), -1)
	}
	return buf
}
