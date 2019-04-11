package errors

import (
	"errors"
	"fmt"
)

// Custom errors
var (
	ErrRepoDirNotFound = errors.New("cannot find repodir")
)

// NoSuchProjectError indicates fail to find project
func NoSuchProjectError(name string) error {
	return fmt.Errorf("cannot find project with name/path '%s'", name)
}

// ProjectNoExistError indicates project not exist error
func ProjectNoExistError(name string) error {
	return fmt.Errorf("project '%s' does not exist", name)
}

// ProjectNotBelongToGroupsError indicates project not belong to gropus
func ProjectNotBelongToGroupsError(name, groups string) error {
	return fmt.Errorf("project '%s' not belong to groups '%s'", name, groups)
}
