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
	Type() string
}

// SSHInfo wraps host and port which ssh_info returned
type SSHInfo struct {
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Type   string `json:"type,omitempty"`
	Expire int64  `json:"expire,omitempty"`
}

// GerritRemote is Gerrit remote server
type GerritRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of GerritRemote
func (v *GerritRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of GerritRemote
func (v *GerritRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// Type returns type of remote
func (v *GerritRemote) Type() string {
	return config.RemoteTypeGerrit
}

// AliRemote is Alibaba remote server
type AliRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of AliRemote
func (v *AliRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of AliRemote
func (v *AliRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// Type returns type of remote
func (v *AliRemote) Type() string {
	return config.RemoteTypeAGit
}

// UnknownRemote is unknown remote server
type UnknownRemote struct {
	manifest.Remote

	SSHInfo *SSHInfo
}

// GetSSHInfo returns SSHInfo field of UnknownRemote
func (v *UnknownRemote) GetSSHInfo() *SSHInfo {
	return v.SSHInfo
}

// GetRemote returns manifest remote field of UnknownRemote
func (v *UnknownRemote) GetRemote() *manifest.Remote {
	return &v.Remote
}

// Type returns type of remote
func (v *UnknownRemote) Type() string {
	return config.RemoteTypeUnknown
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
		remote = &AliRemote{
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
