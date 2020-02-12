package project

import (
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"sync"

	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/helper"
	"github.com/aliyun/git-repo-go/manifest"
	log "github.com/jiangxin/multi-log"
)

// Remote wraps manifest remote.
type Remote struct {
	manifest.Remote
	helper.ProtoHelper

	initialized bool
}

// Initialized indicates whether Remote is initialized.
func (v *Remote) Initialized() bool {
	return v.initialized
}

// ProtoHelperReady indicates ProtoHelper is ready for remote.
func (v *Remote) ProtoHelperReady() bool {
	return v.Initialized() && v.GetType() != ""
}

// NewRemote return new remote object.
func NewRemote(r *manifest.Remote, h helper.ProtoHelper) *Remote {
	if h == nil {
		h = &helper.DefaultProtoHelper{}
	}
	remote := Remote{
		Remote:      *r,
		ProtoHelper: h,

		initialized: true,
	}

	return &remote
}

// GetRemotePushNameURL returns remote name and URL for push.
func (v *Project) GetRemotePushNameURL(remote *Remote) (string, string) {
	var (
		pushURL    string
		defaultURL string
		name       string
	)

	pushURL = v.GetRemotePushURL(remote)
	if pushURL == "" {
		return "", ""
	}
	defaultURL = remote.PushURL
	if defaultURL == "" {
		defaultURL = remote.Fetch
	}
	if strings.TrimSuffix(pushURL, ".git") == strings.TrimSuffix(defaultURL, ".git") {
		name = remote.Name
	}

	return name, pushURL
}

// GetRemotePushURL returns URL for push.
func (v *Project) GetRemotePushURL(remote *Remote) string {
	var (
		sshURL     string
		defaultURL string
	)

	if remote == nil {
		return ""
	}

	sshInfo := remote.GetSSHInfo()
	if sshInfo == nil {
		log.Errorf("cannot get ssh_info for remote: %s", remote.Name)
		return ""
	}
	if sshInfo.Host != "" {
		login := sshInfo.User
		if login == "<email>" {
			login = helper.GetLoginFromEmail(v.UserEmail())
		} else if login == "<login>" {
			u, err := user.Current()
			if err == nil {
				login = u.Username
			}
		}
		if login == "" {
			login = "git"
		}
		sshURL = fmt.Sprintf("ssh://%s@%s", login, sshInfo.Host)
		if sshInfo.Port > 0 && sshInfo.Port != 22 {
			sshURL += ":" + strconv.Itoa(sshInfo.Port)
		}
		sshURL += "/" + v.Name + ".git"
		return sshURL
	}

	defaultURL = remote.PushURL
	if defaultURL == "" {
		defaultURL = remote.Fetch
	}
	return defaultURL
}

// GetDefaultRemote gets default remote for project.
func (v *Project) GetDefaultRemote(useOrigin bool) *Remote {
	return v.GetBranchRemote("", useOrigin)
}

// GetBranchRemote gets default remote for branch of the project.
func (v *Project) GetBranchRemote(branch string, useOrigin bool) *Remote {
	remoteName := v.TrackRemote(branch)
	if remoteName != "" {
		return v.Remotes.Get(remoteName)
	}
	return v.Remotes.Default(useOrigin)
}

// RemoteMap holds all remotes of a project.
type RemoteMap struct {
	remotes       map[string]*Remote
	alias         map[string]string
	defaultRemote string
	lock          sync.RWMutex
}

// Get returns Remote with matched name.
func (v *RemoteMap) Get(name string) *Remote {
	v.lock.RLock()
	r := v.remotes[name]
	if r == nil {
		name = v.alias[name]
		if name != "" {
			r = v.remotes[name]
		}
	}
	v.lock.RUnlock()
	return r
}

// Add will add Remote to map.
func (v *RemoteMap) Add(remote *Remote) *Remote {
	if r := v.Get(remote.Name); r != nil {
		if r.Fetch != remote.Fetch {
			log.Warnf("unmatched fetch url for remote %s: %s != %s",
				remote.Name,
				r.Fetch,
				remote.Fetch,
			)
		}
	} else if remote.Alias != "" {
		if r := v.Get(remote.Alias); r != nil {
			if r.Fetch != remote.Fetch {
				log.Warnf("unmatched fetch url for remote alias %s: %s != %s",
					remote.Name,
					r.Fetch,
					remote.Fetch,
				)
			}
		}
	}

	v.lock.Lock()
	name := remote.Name
	v.remotes[name] = remote
	if remote.Alias != "" {
		v.alias[remote.Alias] = name
	}
	v.lock.Unlock()
	return remote
}

// Default gets default remote.
func (v *RemoteMap) Default(useOrigin bool) *Remote {
	if v.defaultRemote != "" {
		return v.Get(v.defaultRemote)
	}
	if len(v.remotes) == 1 {
		for _, r := range v.remotes {
			return r
		}
	}
	if useOrigin {
		remote := v.Get("origin")
		if remote != nil {
			log.Warnf("multiple remotes are defined, fallback to origin")
			return remote
		}
	}
	return nil
}

// SetDefault sets default remote name.
func (v *RemoteMap) SetDefault(name string) {
	v.defaultRemote = name
}

// NewRemoteMap returns new RemoteMap object.
func NewRemoteMap() *RemoteMap {
	r := RemoteMap{}
	r.remotes = make(map[string]*Remote)
	r.alias = make(map[string]string)
	return &r
}

// LoadRemotes reads git config to load remotes.
func (v *Project) LoadRemotes(remoteMap *RemoteMap, noCache bool) {
	var (
		wg sync.WaitGroup
	)

	v.Remotes = NewRemoteMap()
	cfg := v.Config()
	for _, name := range cfg.Sections() {
		if !strings.HasPrefix(name, "remote.") {
			continue
		}

		wg.Add(1)
		name = strings.TrimPrefix(name, "remote.")
		func(name string) {
			defer wg.Done()
			URL := cfg.Get("remote." + name + ".url")
			if URL == "" {
				log.Warnf("no URL defined for remote: %s", name)
				return
			}
			pushURL := cfg.Get("remote." + name + ".pushurl")
			reviewURL := cfg.Get("remote." + name + ".review")
			protoType := cfg.Get("remote." + name + ".type")

			mr := manifest.Remote{
				Name:    name,
				Fetch:   URL,
				PushURL: pushURL,
				Type:    protoType,
			}

			if reviewURL == "" {
				if pushURL == "" {
					pushURL = URL
				}
				gitURL := config.ParseGitURL(pushURL)
				if gitURL == nil {
					log.Warnf("fail to parse remote: %s, URL: %s", name, pushURL)
				} else {
					reviewURL = gitURL.GetRootURL()
					if reviewURL == "" {
						log.Debugf("cannot get review URL from remote: %s, URL: %s", name, pushURL)
					}
				}
			}
			mr.Review = reviewURL
			if remoteMap != nil {
				if r := remoteMap.Get(name); r != nil {
					if r.Review != "" {
						mr.Review = r.Review
					}
					if r.Revision != "" {
						mr.Revision = r.Revision
					}
					if r.Type != "" {
						mr.Type = r.Type
					}
					newRemote := NewRemote(&mr, r.ProtoHelper)
					v.Remotes.Add(newRemote)
					return
				}
			}
			v.AddRemote(&mr, noCache)
		}(name)
	}

	wg.Wait()

	// If project is uninitialized, add manifest remote.
	if v.ManifestRemote != nil {
		remoteName := v.ManifestRemote.Name
		if v.Remotes.Get(remoteName) == nil {
			if r := remoteMap.Get(remoteName); r != nil {
				newRemote := NewRemote(v.ManifestRemote, r.ProtoHelper)
				v.Remotes.Add(newRemote)
			}
		}
	}
}

// AddRemote will add new Remote to project's Remotes.
func (v *Project) AddRemote(mr *manifest.Remote, noCache bool) *Remote {
	var (
		protoHelper helper.ProtoHelper
	)

	reviewURL := mr.Review
	if reviewURL != "" {
		if mr.Type != "" {
			sshInfo := &helper.SSHInfo{ProtoType: mr.Type}
			protoHelper = helper.NewProtoHelper(sshInfo)
		} else {
			query := helper.NewSSHInfoQuery(v.SSHInfoCacheFile())
			sshInfo, err := query.GetSSHInfo(mr.Review, !noCache)
			if err != nil {
				log.Debug(err)
			} else {
				protoHelper = helper.NewProtoHelper(sshInfo)
			}
		}
	}
	remote := NewRemote(mr, protoHelper)
	return v.Remotes.Add(remote)
}
