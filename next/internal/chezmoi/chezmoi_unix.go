// +build !windows

package chezmoi

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

var umask os.FileMode

//nolint:gochecknoinits
func init() {
	umask = os.FileMode(syscall.Umask(0))
	syscall.Umask(int(umask))
}

// GetUmask returns the umask.
func GetUmask() os.FileMode {
	return umask
}

// TrimDirPrefix returns path with the directory prefix dir stripped. path must
// be an absolute path with forward slashes.
func TrimDirPrefix(path, dir string) (string, error) {
	// FIXME add absSlash check
	prefix := dir + "/"
	if !strings.HasPrefix(path, prefix) {
		return "", fmt.Errorf("%q dpes not have dir prefix %q", path, dir)
	}
	return path[len(prefix):], nil
}

// SetUmask sets the umask.
func SetUmask(newUmask os.FileMode) {
	umask = newUmask
	syscall.Umask(int(umask))
}

// isExecutable returns if info is executable.
func isExecutable(info os.FileInfo) bool {
	return info.Mode().Perm()&0o111 != 0
}

// isPrivate returns if info is private.
func isPrivate(info os.FileInfo) bool {
	return info.Mode().Perm()&0o77 == 0
}

// umaskPermEqual returns if two permissions are equal after applying umask.
func umaskPermEqual(perm1, perm2, umask os.FileMode) bool {
	return perm1&^umask == perm2&^umask
}
