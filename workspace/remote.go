package workspace

import (
	"net/http"

	"code.alibaba-inc.com/force/git-repo/helper"
	"code.alibaba-inc.com/force/git-repo/project"
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
		query *helper.SSHInfoQuery
	)

	if v.Manifest == nil || v.Manifest.Remotes == nil {
		return nil
	}

	query = helper.NewSSHInfoQuery(v.ManifestProject.ProtoCacheFile())
	for _, r := range v.Manifest.Remotes {
		sshInfo, err := query.GetSSHInfo(r.Review, !noCache)
		if err != nil {
			return err
		}
		protoHelper := helper.NewProtoHelper(sshInfo)
		remote := project.NewRemote(&r, protoHelper)
		v.RemoteMap[r.Name] = *remote
	}

	for i := range v.Projects {
		name := v.Projects[i].ManifestRemote.Name
		if name == "" {
			log.Warnf("empty remote for project '%s'",
				v.Projects[i].Name)
			continue
		}
		if _, ok := v.RemoteMap[name]; ok {
			v.Projects[i].Remote = v.RemoteMap[name]
		}
	}

	return nil
}
