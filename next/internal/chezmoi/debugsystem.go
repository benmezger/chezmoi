package chezmoi

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	vfs "github.com/twpayne/go-vfs"
)

// A DebugSystem wraps a System and logs all of the actions it executes.
type DebugSystem struct {
	s      System
	ps     PersistentState
	logger zerolog.Logger
}

// NewDebugSystem returns a new DebugSystem.
func NewDebugSystem(system System, logger zerolog.Logger) *DebugSystem {
	return &DebugSystem{
		s:      system,
		ps:     newDebugPersistentState(system.PersistentState(), logger),
		logger: logger,
	}
}

// Chmod implements System.Chmod.
func (s *DebugSystem) Chmod(name string, mode os.FileMode) error {
	err := s.s.Chmod(name, mode)
	s.logger.Debug().
		Str("name", name).
		Int("mode", int(mode)).
		Err(err).
		Msg("Chmod")
	return err
}

// Glob implements System.Glob.
func (s *DebugSystem) Glob(name string) ([]string, error) {
	matches, err := s.s.Glob(name)
	s.logger.Debug().
		Str("name", name).
		Strs("matches", matches).
		Err(err).
		Msg("Glob")
	return matches, err
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *DebugSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	type result struct {
		startTime time.Time
		output    []byte
		err       error
	}

	resultCh := make(chan result)
	go func(resultCh chan<- result) {
		defer close(resultCh)
		start := time.Now()
		output, err := s.s.IdempotentCmdOutput(cmd)
		resultCh <- result{
			startTime: start,
			output:    output,
			err:       err,
		}
	}(resultCh)

	var r result
	select {
	case r = <-resultCh:
	case <-time.After(1 * time.Second):
		s.logger.Debug().
			Dict("cmd", cmdDict(cmd)).
			Msg("IdempotentCmdOutput")
		r = <-resultCh
	}

	s.logger.Debug().
		Dict("cmd", cmdDict(cmd)).
		Str("output", firstFewBytes(r.output)).
		Err(r.err).
		Dur("duration", time.Since(r.startTime)).
		Msg("IdempotentCmdOutput")

	return r.output, r.err
}

// Lstat implements System.Lstat.
func (s *DebugSystem) Lstat(name string) (os.FileInfo, error) {
	info, err := s.s.Lstat(name)
	s.logger.Debug().
		Str("name", name).
		Err(err).
		Msg("Lstat")
	return info, err
}

// Mkdir implements System.Mkdir.
func (s *DebugSystem) Mkdir(name string, perm os.FileMode) error {
	err := s.s.Mkdir(name, perm)
	s.logger.Debug().
		Str("name", name).
		Int("perm", int(perm)).
		Err(err).
		Msg("Mkdir")
	return err
}

// PersistentState implements System.PersistentState.
func (s *DebugSystem) PersistentState() PersistentState {
	return s.ps
}

// RawPath implements System.RawPath.
func (s *DebugSystem) RawPath(path string) (string, error) {
	return s.s.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *DebugSystem) ReadDir(name string) ([]os.FileInfo, error) {
	infos, err := s.s.ReadDir(name)
	s.logger.Debug().
		Str("name", name).
		Err(err).
		Msg("ReadDir")
	return infos, err
}

// ReadFile implements System.ReadFile.
func (s *DebugSystem) ReadFile(filename string) ([]byte, error) {
	data, err := s.s.ReadFile(filename)
	s.logger.Debug().
		Str("filename", filename).
		Str("data", firstFewBytes(data)).
		Err(err).
		Msg("ReadFile")
	return data, err
}

// Readlink implements System.Readlink.
func (s *DebugSystem) Readlink(name string) (string, error) {
	linkname, err := s.s.Readlink(name)
	s.logger.Debug().
		Str("name", name).
		Str("linkname", linkname).
		Err(err).
		Msg("Readlink")
	return linkname, err
}

// RemoveAll implements System.RemoveAll.
func (s *DebugSystem) RemoveAll(name string) error {
	err := s.s.RemoveAll(name)
	s.logger.Debug().
		Str("name", name).
		Err(err).
		Msg("RemoveAll")
	return err
}

// Rename implements System.Rename.
func (s *DebugSystem) Rename(oldpath, newpath string) error {
	err := s.s.Rename(oldpath, newpath)
	s.logger.Debug().
		Str("oldpath", oldpath).
		Str("newpath", newpath).
		Err(err).
		Msg("Rename")
	return err
}

// RunCmd implements System.RunCmd.
func (s *DebugSystem) RunCmd(cmd *exec.Cmd) error {
	type result struct {
		startTime time.Time
		err       error
	}

	resultCh := make(chan result)
	go func(resultCh chan<- result) {
		defer close(resultCh)
		start := time.Now()
		err := s.s.RunCmd(cmd)
		resultCh <- result{
			startTime: start,
			err:       err,
		}
	}(resultCh)

	var r result
	select {
	case r = <-resultCh:
	case <-time.After(1 * time.Second):
		s.logger.Debug().
			Dict("cmd", cmdDict(cmd)).
			Msg("RunCmd")
		r = <-resultCh
	}

	s.logger.Debug().
		Dict("cmd", cmdDict(cmd)).
		Err(r.err).
		Dur("duration", time.Since(r.startTime)).
		Msg("RunCmd")

	return r.err
}

// RunScript implements System.RunScript.
func (s *DebugSystem) RunScript(scriptname, dir string, data []byte) error {
	type result struct {
		startTime time.Time
		err       error
	}

	resultCh := make(chan result)
	go func(resultCh chan<- result) {
		defer close(resultCh)
		start := time.Now()
		err := s.s.RunScript(scriptname, dir, data)
		resultCh <- result{
			startTime: start,
			err:       err,
		}
	}(resultCh)

	var r result
	select {
	case r = <-resultCh:
	case <-time.After(1 * time.Second):
		s.logger.Debug().
			Str("scriptname", scriptname).
			Str("dir", dir).
			Str("data", firstFewBytes(data)).
			Msg("RunScript")
		r = <-resultCh
	}

	s.logger.Debug().
		Str("scriptname", scriptname).
		Str("dir", dir).
		Str("data", firstFewBytes(data)).
		Err(r.err).
		Dur("duration", time.Since(r.startTime)).
		Msg("RunScript")

	return r.err
}

// Stat implements System.Stat.
func (s *DebugSystem) Stat(name string) (os.FileInfo, error) {
	info, err := s.s.Stat(name)
	s.logger.Debug().
		Str("name", name).
		Err(err).
		Msg("Stat")
	return info, err
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DebugSystem) UnderlyingFS() vfs.FS {
	return s.s.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *DebugSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	err := s.s.WriteFile(name, data, perm)
	s.logger.Debug().
		Str("name", name).
		Str("data", firstFewBytes(data)).
		Int("perm", int(perm)).
		Err(err).
		Msg("WriteFile")
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *DebugSystem) WriteSymlink(oldname, newname string) error {
	err := s.s.WriteSymlink(oldname, newname)
	s.logger.Debug().
		Str("oldname", oldname).
		Str("newname", newname).
		Err(err).
		Msg("WriteSymlink")
	return err
}

func cmdDict(cmd *exec.Cmd) *zerolog.Event {
	return zerolog.Dict().
		Str("path", cmd.Path).
		Strs("args", cmd.Args)
}

// firstFewBytes returns the first few bytes of data in a human-readable form.
func firstFewBytes(data []byte) string {
	const few = 64
	if len(data) > few {
		data = append([]byte{}, data[:few]...)
		data = append(data, '.', '.', '.')
	}
	s := strconv.Quote(string(data))
	return s[1 : len(s)-1]
}
