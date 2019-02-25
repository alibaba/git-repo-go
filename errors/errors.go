package errors

import "errors"

// Custom errors
var (
	ErrRepoDirNotFound = errors.New("cannot find repodir")
)
