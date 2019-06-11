package project

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/manifest"
)

var (
	sshInfoPattern = regexp.MustCompile(`^[\S]+ [0-9]+$`)
)

// Remote interface wraps remote server of Gerrit or Alibaba
type Remote interface {
	GetSSHInfo() *SSHInfo
	GetRemote() *manifest.Remote
	GetType() string
	UploadCommands(o *UploadOptions, branch *ReviewableBranch) ([]string, error)
}

// RemoteWithError wraps Remote and Error when parsing remote
type RemoteWithError struct {
	Remote Remote
	Error  error
}

// SSHInfo wraps host and port which ssh_info returned
type SSHInfo struct {
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Type   string `json:"type,omitempty"`
	Expire int64  `json:"expire,omitempty"`
}

// Strings returns "<Host><SP><Port>"
func (v SSHInfo) String() string {
	if v.Host == "" || v.Port == 0 {
		return ""
	}
	return fmt.Sprintf("%s %d", v.Host, v.Port)
}

func newSSHInfo(data string) (*SSHInfo, error) {
	var (
		host  string
		port  int
		err   error
		items []string
	)

	items = strings.SplitN(data, " ", 2)
	if len(items) != 2 {
		return nil, fmt.Errorf("wrong ssh_info format: %s", data)
	}

	host = items[0]
	port, err = strconv.Atoi(items[1])
	if err != nil {
		return nil, fmt.Errorf("bad port number (%s) in ssh_info: %s", items[1], err)
	}

	return &SSHInfo{
		Host: host,
		Port: port,
	}, nil
}

// NewRemote parses ssh_info and return Remote interface
func NewRemote(r *manifest.Remote, remoteType, data string) (Remote, error) {
	var (
		err     error
		remote  Remote
		sshInfo *SSHInfo
	)

	data = strings.TrimSpace(data)
	if data != "" {
		if sshInfoPattern.MatchString(data) {
			sshInfo, err = newSSHInfo(data)
			if err != nil {
				return nil, err
			}
			if remoteType == "" {
				remoteType = config.RemoteTypeGerrit
			}
		} else {
			sshInfo = &SSHInfo{}
			err = json.Unmarshal([]byte(data), sshInfo)
			if err != nil {
				return nil, fmt.Errorf("fail to parse json from: %s", data)
			}
		}
	}

	if sshInfo != nil && sshInfo.Type != "" {
		remoteType = sshInfo.Type
	}

	switch strings.ToLower(remoteType) {
	case config.RemoteTypeGerrit:
		remote = &GerritRemote{
			Remote:  *r,
			SSHInfo: sshInfo,
		}
	case config.RemoteTypeAGit:
		remote = &AGitRemote{
			Remote:  *r,
			SSHInfo: sshInfo,
		}
	default:
		remote = &UnknownRemote{
			Remote:  *r,
			SSHInfo: sshInfo,
		}
	}

	return remote, nil
}
