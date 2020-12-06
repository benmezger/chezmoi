package chezmoi

import (
	"fmt"
	"os"
	"strings"
)

// GetUmask returns the umask.
func GetUmask() os.FileMode {
	return os.FileMode(0)
}

// TrimDirPrefix returns path with the directory prefix dir stripped. path must
// be an absolute path with forward slashes.
func TrimDirPrefix(path, dir string) (string, error) {
	prefix := strings.ToLower(dir + "/")
	if !strings.HasPrefix(strings.ToLower(path), prefix) {
		return "", fmt.Errorf("%q does not have dir prefix %q", path, dir)
	}
	return path[len(prefix):], nil
}

// SetUmask sets the umask.
func SetUmask(umask os.FileMode) {}

// isExecutable returns false on Windows.
func isExecutable(info os.FileInfo) bool {
	return false
}

// permIsPrivate returns false on Windows.
func isPrivate(info os.FileInfo) bool {
	return false
}

// umaskPermEqual returns true on Windows.
func umaskPermEqual(perm1 os.FileMode, perm2 os.FileMode, umask os.FileMode) bool {
	return true
}
