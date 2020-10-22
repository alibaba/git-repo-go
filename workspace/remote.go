package workspace

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alibaba/git-repo-go/helper"
	"github.com/alibaba/git-repo-go/project"
	log "github.com/jiangxin/multi-log"
)

const (
	remoteCallTimeout = 10
)

var (
	httpClient *http.Client
)

// LoadRemotes calls remote API to get server type and other info.
func (v *RepoWorkSpace) LoadRemotes(noCache bool) error {
	var (
		query         *helper.SSHInfoQuery
		remoteMap     = project.NewRemoteMap()
		failedRemotes = []string{}
	)

	if v.Manifest == nil || v.Manifest.Remotes == nil {
		return nil
	}

	query = helper.NewSSHInfoQuery(v.ManifestProject.SSHInfoCacheFile())
	for _, r := range v.Manifest.Remotes {
		var (
			sshInfo *helper.SSHInfo
			err     error
		)

		if r.Review == "" {
			log.Infof("attribute 'review' is not defined in remote '%s'", r.Name)
			sshInfo = &helper.SSHInfo{}
		} else {
			sshInfo, err = query.GetSSHInfo(r.Review, !noCache)
			if err != nil {
				log.Error(err)
				failedRemotes = append(failedRemotes, r.Name)
				continue
			}
		}
		protoHelper := helper.NewProtoHelper(sshInfo)
		remote := project.NewRemote(&r, protoHelper)
		remoteMap.Add(remote)
	}

	for i := range v.Projects {
		v.Projects[i].LoadRemotes(remoteMap, noCache)

		name := v.Projects[i].ManifestRemote.Name
		if name == "" {
			log.Warnf("empty remote for project '%s'",
				v.Projects[i].Name)
		} else {
			v.Projects[i].Remotes.SetDefault(name)
		}
	}

	if len(failedRemotes) == 0 {
		return nil
	}
	return fmt.Errorf("fail to load remotes: %s", strings.Join(failedRemotes, ", "))
}
