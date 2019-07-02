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

// Package errors implements pre-defined errors.
package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrRepoDirNotFound indicates fail to find .repo dir
	ErrRepoDirNotFound = errors.New("cannot find repodir")
)

// NoSuchProjectError indicates fail to find project.
func NoSuchProjectError(name string) error {
	return fmt.Errorf("cannot find project with name/path '%s'", name)
}

// ProjectNoExistError indicates project not exist error.
func ProjectNoExistError(name string) error {
	return fmt.Errorf("project '%s' does not exist", name)
}

// ProjectNotBelongToGroupsError indicates project not belong to gropus.
func ProjectNotBelongToGroupsError(name, groups string) error {
	return fmt.Errorf("project '%s' not belong to groups '%s'", name, groups)
}
