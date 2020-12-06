// +build !windows

package chezmoi

import (
	"errors"
	"os"
	"path"
	"syscall"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
	"go.uber.org/multierr"
)

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	// Special case: if writing to the real filesystem, use
	// github.com/google/renameio.
	if s.FS == vfs.OSFS {
		dir := path.Dir(filename)
		dev, ok := s.devCache[dir]
		if !ok {
			info, err := s.Stat(dir)
			if err != nil {
				return err
			}
			statT, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return errors.New("os.FileInfo.Sys() cannot be converted to a *syscall.Stat_t")
			}
			dev = uint(statT.Dev)
			s.devCache[dir] = dev
		}
		tempDir, ok := s.tempDirCache[dev]
		if !ok {
			tempDir = renameio.TempDir(dir)
			s.tempDirCache[dev] = tempDir
		}
		t, err := renameio.TempFile(tempDir, filename)
		if err != nil {
			return err
		}
		defer func() {
			_ = t.Cleanup()
		}()
		if err := t.Chmod(perm); err != nil {
			return err
		}
		if _, err := t.Write(data); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}

	return writeFile(s.FS, filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname, newname string) error {
	// Special case: if writing to the real filesystem, use
	// github.com/google/renameio.
	if s.FS == vfs.OSFS {
		return renameio.Symlink(oldname, newname)
	}
	if err := s.FS.RemoveAll(newname); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.FS.Symlink(oldname, newname)
}

// writeFile is like ioutil.writeFile but always sets perm before writing data.
// ioutil.writeFile only sets the permissions when creating a new file. We need
// to ensure permissions, so we use our own implementation.
func writeFile(fs vfs.FS, filename string, data []byte, perm os.FileMode) (err error) {
	// Create a new file, or truncate any existing one.
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, f.Close())
	}()

	// Set permissions after truncation but before writing any data, in case the
	// file contained private data before, but before writing the new contents,
	// in case the contents contain private data after.
	if err = f.Chmod(perm); err != nil {
		return
	}

	_, err = f.Write(data)
	return
}
