package chezmoi

import (
	"os"
)

// An ActualStateEntry represents the actual state of an entry in the
// filesystem.
type ActualStateEntry interface {
	Path() string
	Remove(s System) error
}

// A ActualStateAbsent represents the absence of an entry in the filesystem.
type ActualStateAbsent struct {
	path string
}

// A ActualStateDir represents the state of a directory in the filesystem.
type ActualStateDir struct {
	path string
	perm os.FileMode
}

// A ActualStateFile represents the state of a file in the filesystem.
type ActualStateFile struct {
	path string
	perm os.FileMode
	*lazyContents
}

// A ActualStateSymlink represents the state of a symlink in the filesystem.
type ActualStateSymlink struct {
	path string
	*lazyLinkname
}

// NewActualStateEntry returns a new ActualStateEntry populated with path from
// fs.
func NewActualStateEntry(s System, path string) (ActualStateEntry, error) {
	info, err := s.Lstat(path)
	switch {
	case os.IsNotExist(err):
		return &ActualStateAbsent{
			path: path,
		}, nil
	case err != nil:
		return nil, err
	}
	//nolint:exhaustive
	switch info.Mode() & os.ModeType {
	case 0:
		return &ActualStateFile{
			path: path,
			perm: info.Mode() & os.ModePerm,
			lazyContents: &lazyContents{
				contentsFunc: func() ([]byte, error) {
					return s.ReadFile(path)
				},
			},
		}, nil
	case os.ModeDir:
		return &ActualStateDir{
			path: path,
			perm: info.Mode() & os.ModePerm,
		}, nil
	case os.ModeSymlink:
		return &ActualStateSymlink{
			path: path,
			lazyLinkname: &lazyLinkname{
				linknameFunc: func() (string, error) {
					return s.Readlink(path)
				},
			},
		}, nil
	default:
		return nil, &unsupportedFileTypeError{
			path: path,
			mode: info.Mode(),
		}
	}
}

// Path returns d's path.
func (d *ActualStateAbsent) Path() string {
	return d.path
}

// Remove removes d.
func (d *ActualStateAbsent) Remove(s System) error {
	return nil
}

// Path returns d's path.
func (d *ActualStateDir) Path() string {
	return d.path
}

// Remove removes d.
func (d *ActualStateDir) Remove(s System) error {
	return s.RemoveAll(d.path)
}

// Path returns d's path.
func (d *ActualStateFile) Path() string {
	return d.path
}

// Remove removes d.
func (d *ActualStateFile) Remove(s System) error {
	return s.RemoveAll(d.path)
}

// Path returns d's path.
func (d *ActualStateSymlink) Path() string {
	return d.path
}

// Remove removes d.
func (d *ActualStateSymlink) Remove(s System) error {
	return s.RemoveAll(d.path)
}
