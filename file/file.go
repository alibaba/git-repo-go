package file

import (
	"errors"
	"os"
)

// Common errors
var (
	ErrNoFileName = errors.New("file name is not set")
)

// File implements a helper for os.File.
type File struct {
	Name string

	perm os.FileMode
}

// New creates File object.
func New(name string) *File {
	return &File{Name: name}
}

// SetName set Name for file.
func (v *File) SetName(name string) *File {
	v.Name = name
	return v
}

// SetPerm sets file permission.
func (v *File) SetPerm(perm os.FileMode) *File {
	v.perm = perm
	return v
}

// SetExecutable sets perm to 0755.
func (v *File) SetExecutable() *File {
	return v.SetPerm(0755)
}

func (v *File) openWrite(flag int) (*os.File, error) {
	if v.Name == "" {
		return nil, ErrNoFileName
	}

	if flag&(os.O_RDWR|os.O_WRONLY) == 0 {
		flag |= os.O_WRONLY
	}

	if v.perm == 0 {
		v.perm = 0644
	}

	return os.OpenFile(v.Name, flag, v.perm)
}

// Open opens file for reading.
func (v *File) Open() (*os.File, error) {
	if v.Name == "" {
		return nil, ErrNoFileName
	}
	return os.Open(v.Name)
}

// OpenReadWrite opens file for read and write.
func (v *File) OpenReadWrite() (*os.File, error) {
	return v.openWrite(os.O_RDWR)
}

// OpenCreateReadWrite opens file for read and write, and create if not exist.
func (v *File) OpenCreateReadWrite() (*os.File, error) {
	return v.openWrite(os.O_RDWR | os.O_CREATE)
}

// OpenCreateReadWriteExcl create file for read and write, and file must not exist.
func (v *File) OpenCreateReadWriteExcl() (*os.File, error) {
	return v.openWrite(os.O_RDWR | os.O_CREATE | os.O_EXCL)
}

// OpenRewrite opens file for write, and will truncate file if already exist.
func (v *File) OpenRewrite() (*os.File, error) {
	return v.openWrite(os.O_TRUNC)
}

// OpenCreateRewrite open file for rewrite, and will create file if not exist.
func (v *File) OpenCreateRewrite() (*os.File, error) {
	return v.openWrite(os.O_CREATE | os.O_TRUNC)
}

// OpenCreateRewriteExcl create file for rewrite, and fail if file exist.
func (v *File) OpenCreateRewriteExcl() (*os.File, error) {
	return v.openWrite(os.O_CREATE | os.O_TRUNC | os.O_EXCL)
}

// OpenAppend open file for append.
func (v *File) OpenAppend() (*os.File, error) {
	return v.openWrite(os.O_APPEND)
}

// OpenCreateAppend open file for append, and create if not exist.
func (v *File) OpenCreateAppend() (*os.File, error) {
	return v.openWrite(os.O_CREATE | os.O_APPEND)
}
