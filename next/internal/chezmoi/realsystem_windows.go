package chezmoi

import (
	"os"
)

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname, newname string) error {
	if err := s.FS.RemoveAll(newname); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.FS.Symlink(oldname, newname)
}
