package chezmoi

import (
	"path/filepath"
)

// An OSPath is a native OS path.
type OSPath struct {
	s string
}

// NewOSPath returns a new OSPath.
func NewOSPath(s string) *OSPath {
	return &OSPath{
		s: filepath.FromSlash(s),
	}
}

// AbsSlash returns p converted to an absolute path with backslashes replaced by
// forward slashes.
func (p *OSPath) AbsSlash() (string, error) {
	abs, err := filepath.Abs(p.s)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(abs), nil
}

// Dir returns p's directory.
func (p *OSPath) Dir() *OSPath {
	return &OSPath{
		s: filepath.Dir(p.s),
	}
}

// Empty returns if p is empty.
func (p *OSPath) Empty() bool {
	return p.s != ""
}

// Join joins elems on to p.
func (p *OSPath) Join(elems ...string) *OSPath {
	return &OSPath{
		s: filepath.Join(append([]string{p.s}, elems...)...),
	}
}

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (p *OSPath) MarshalText() ([]byte, error) {
	return []byte(p.s), nil
}

func (p *OSPath) String() string {
	return p.s
}

// UnmarshalText implements encoding.TextUnmarshaler.UnmarshalText.
func (p *OSPath) UnmarshalText(data []byte) error {
	p.s = string(data)
	return nil
}
