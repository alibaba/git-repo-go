package workspace

import (
	"net/http"

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
		query     *helper.SSHInfoQuery
		remoteMap = project.NewRemoteMap()
	)

	if v.Manifest == nil || v.Manifest.Remotes == nil {
		return nil
	}

	query = helper.NewSSHInfoQuery(v.ManifestProject.SSHInfoCacheFile())
	for _, r := range v.Manifest.Remotes {
		sshInfo, err := query.GetSSHInfo(r.Review, !noCache)
		if err != nil {
			return err
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

	return nil
}
