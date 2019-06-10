package project

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"code.alibaba-inc.com/force/git-repo/color"
	log "github.com/jiangxin/multi-log"
)

type gitStatus struct {
	SrcMode  string
	DestMode string
	SrcOID   string
	DestOID  string
	Status   string
	Level    string
	SrcPath  string
	Path     string
}

func parseGitStatus(out []byte) []*gitStatus {
	var (
		wantMode    = true
		wantPath    bool
		wantTwoPath bool
		result      []*gitStatus
		s           *gitStatus
	)
	for _, buf := range bytes.Split(out, []byte("\x00")) {
		if len(buf) == 0 {
			break
		}
		if wantMode {
			wantMode = false
			wantPath = true
			s = &gitStatus{}
			if buf[0] != ':' {
				log.Errorf("status not starts with colon: %s", string(buf))
				break
			}
			buf = buf[1:]
			fields := bytes.Split(buf, []byte(" "))
			if len(fields) != 5 {
				log.Errorf("wrong status line: %s", string(buf))
				break
			}
			s.SrcMode = string(fields[0])
			s.DestMode = string(fields[1])
			s.SrcOID = string(fields[2])
			s.DestOID = string(fields[3])
			s.Status = string(fields[4])
			if s.Status[0] == 'C' || s.Status[0] == 'R' {
				s.Level = string(fields[4][1:])
				s.Status = string(fields[4][0])
				wantTwoPath = true
			} else {
				wantTwoPath = false
			}
			continue
		}

		if wantPath {
			wantPath = false
			s.Path = string(buf)
			if wantTwoPath {
				continue
			} else {
				wantMode = true
			}
		}

		if wantTwoPath {
			wantTwoPath = false
			wantMode = true
			s.SrcPath = s.Path
			s.Path = string(buf)
		}

		if wantMode {
			result = append(result, s)
		}
	}
	return result
}

func combineGitStatus(sti, stf []*gitStatus, sto []string) string {
	var (
		stiMap = make(map[string]*gitStatus)
		stfMap = make(map[string]*gitStatus)
		keys   []string
		result string
	)
	for _, s := range sti {
		stiMap[s.Path] = s
		keys = append(keys, s.Path)
	}
	for _, s := range stf {
		stfMap[s.Path] = s
		if _, ok := stiMap[s.Path]; !ok {
			keys = append(keys, s.Path)
		}
	}
	for _, s := range sto {
		keys = append(keys, string(s))
	}
	sort.Strings(keys)
	for _, p := range keys {
		i := stiMap[p]
		f := stfMap[p]
		iStatus := ""
		fStatus := ""
		line := ""
		if i != nil {
			iStatus = strings.ToUpper(i.Status)
		} else {
			iStatus = "-"
		}
		if f != nil {
			fStatus = strings.ToLower(f.Status)
		} else {
			fStatus = "-"
		}
		if i != nil && i.SrcPath != "" {
			line = fmt.Sprintf(" %s%s\t%s => %s (%s%%)",
				iStatus,
				fStatus,
				i.SrcPath,
				i.Path,
				i.Level,
			)
		} else {
			line = fmt.Sprintf(" %s%s\t%s", iStatus, fStatus, p)
		}

		if i != nil && f == nil {
			// add new entry
			line = color.Color("green", "", "") + line + color.Reset()
		} else if f != nil {
			// changed entry
			line = color.Color("red", "", "") + line + color.Reset()
		} else {
			// untracked
			line = color.Color("red", "", "") + line + color.Reset()
		}
		result += line + "\n"
	}
	return result
}

// Status shows combined output of git status for project.
func (v Project) Status() *CmdExecResult {
	result := NewCmdExecResult(&v)

	rb := v.IsRebaseInProgress()
	if rb {
		result.Error = fmt.Errorf("prior sync failed; rebase still in progress")
		return result
	}

	di := v.ExecuteCommand("git",
		"diff-index",
		"-z",
		"-M",
		"--cached",
		"HEAD")
	df := v.ExecuteCommand("git",
		"diff-files",
		"-z")
	do := v.ExecuteCommand("git",
		"ls-files",
		"-z",
		"--others",
		"--exclude-standard")

	if di.Empty() && df.Empty() && do.Empty() {
		return result
	}

	sti := parseGitStatus(di.Out)
	stf := parseGitStatus(df.Out)
	sto := []string{}
	for _, name := range bytes.Split(do.Out, []byte("\x00")) {
		if len(name) > 0 {
			sto = append(sto, string(name))
		}
	}

	output := combineGitStatus(sti, stf, sto)
	result.Out = []byte(output)

	if di.Error != nil && df.Error != nil {
		errMsg := di.Stderr()
		if errMsg != "" && errMsg[len(errMsg)-1] != '\n' {
			errMsg += "\n"
		}
		errMsg += df.Stderr()
		result.Error = errors.New(errMsg)
	} else if di.Error != nil {
		result.Error = di.Error
	} else if df.Error != nil {
		result.Error = df.Error
	}

	return result
}
