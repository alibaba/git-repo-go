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

package helper

import (
	"encoding/json"
	"io"
	"strings"

	"code.alibaba-inc.com/force/git-repo/project"
)

// GitPushCommand holds command and args for git command.
type GitPushCommand struct {
	Cmd       string   `json:"cmd,omitempty"`
	Args      []string `json:"args,omitempty"`
	Env       []string `json:"env,omitempty"`
	GitConfig []string `json:"gitconfig,omitempty"`
}

// ProtoHelper defines interface for proto helper.
type ProtoHelper interface {
	GetType() string
	GetGitPushCommandPipe(io.Reader) ([]byte, error)
	GetGitPushCommand(*project.UploadOptions) (*GitPushCommand, error)
	GetDownloadRef(string, string) (string, error)
}

// NewProtoHelper returns proto helper for specific proto type.
func NewProtoHelper(protoType string) ProtoHelper {
	protoType = strings.ToLower(protoType)
	switch protoType {
	case "agit":
		return &AGitHelper{}
	case "gerrit":
		return &GerritHelper{}
	}
	return &ExternalHelper{ProtoType: protoType}
}

func getGitPushCommandPipe(proto ProtoHelper, reader io.Reader) ([]byte, error) {
	var (
		o   = project.UploadOptions{}
		err error
	)

	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&o)
	if err != nil {
		return nil, err
	}

	cmd, err := proto.GetGitPushCommand(&o)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&cmd)
}
