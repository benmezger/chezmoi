package chezmoi

import (
	"os"

	"github.com/rs/zerolog"
)

// A PersistentState is a persistent state.
type PersistentState interface {
	Get(bucket, key []byte) ([]byte, error)
	Delete(bucket, key []byte) error
	ForEach(bucket []byte, fn func(k, v []byte) error) error
	OpenOrCreate() error
	Set(bucket, key, value []byte) error
}

// A debugPersistentState wraps a PersistentState and logs to a log.Logger.
type debugPersistentState struct {
	s      PersistentState
	logger zerolog.Logger
}

// A dryRunPersistentState wraps a PersistentState and drops all writes but
// records that they occurred.
//
// FIXME mock writes (e.g. writes should not affect the underlying
// PersistentState but subsequent reads should return as if the write occurred).
type dryRunPersistentState struct {
	s        PersistentState
	modified bool
}

// A nullPersistentState is an empty PersistentState that returns the zero value
// for all reads and silently consumes all writes.
type nullPersistentState struct{}

// A readOnlyPersistentState wraps a PeristentState but returns an error on any
// write.
type readOnlyPersistentState struct {
	s PersistentState
}

// newDebugPersistentState returns a new debugPersistentState that wraps s and
// logs to logger.
func newDebugPersistentState(s PersistentState, logger zerolog.Logger) *debugPersistentState {
	return &debugPersistentState{
		s:      s,
		logger: logger,
	}
}

// Delete implements PersistentState.Delete.
func (s *debugPersistentState) Delete(bucket, key []byte) error {
	err := s.s.Delete(bucket, key)
	s.logger.Debug().
		Bytes("bucket", bucket).
		Bytes("key", key).
		Err(err).
		Msg("Delete")
	return err
}

// ForEach implements PersistentState.ForEach.
func (s *debugPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	err := s.s.ForEach(bucket, func(k, v []byte) error {
		err := fn(k, v)
		s.logger.Debug().
			Bytes("bucket", bucket).
			Bytes("key", k).
			Bytes("value", v).
			Err(err).
			Msg("ForEach")
		return err
	})
	s.logger.Debug().
		Bytes("bucket", bucket).
		Err(err)
	return err
}

// Get implements PersistentState.Get.
func (s *debugPersistentState) Get(bucket, key []byte) ([]byte, error) {
	value, err := s.s.Get(bucket, key)
	s.logger.Debug().
		Bytes("bucket", bucket).
		Bytes("key", key).
		Bytes("value", value).
		Err(err).
		Msg("Get")
	return value, err
}

// OpenOrCreate implements PersistentState.OpenOrCreate.
func (s *debugPersistentState) OpenOrCreate() error {
	err := s.s.OpenOrCreate()
	s.logger.Debug().
		Err(err).
		Msg("OpenOrCreate")
	return err
}

// Set implements PersistentState.Set.
func (s *debugPersistentState) Set(bucket, key, value []byte) error {
	err := s.s.Set(bucket, key, value)
	s.logger.Debug().
		Bytes("bucket", bucket).
		Bytes("key", key).
		Bytes("value", value).
		Err(err).
		Msg("Set")
	return err
}

// newDryRunPersistentState returns a new dryRunPersistentState that wraps s.
func newDryRunPersistentState(s PersistentState) *dryRunPersistentState {
	return &dryRunPersistentState{
		s: s,
	}
}

// Get implements PersistentState.Get.
func (s *dryRunPersistentState) Get(bucket, key []byte) ([]byte, error) {
	return s.s.Get(bucket, key)
}

// Delete implements PersistentState.Delete.
func (s *dryRunPersistentState) Delete(bucket, key []byte) error {
	s.modified = true
	return nil
}

// ForEach implements PersistentState.ForEach.
func (s *dryRunPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	return s.s.ForEach(bucket, fn)
}

// OpenOrCreate implements PersistentState.OpenOrCreate.
func (s *dryRunPersistentState) OpenOrCreate() error {
	s.modified = true // FIXME this will give false negatives if s.s already exists, need to separate create from open
	return s.s.OpenOrCreate()
}

// Set implements PersistentState.Set.
func (s *dryRunPersistentState) Set(bucket, key, value []byte) error {
	s.modified = true
	// FIXME do we need to remember that the value has been set?
	return nil
}

func (nullPersistentState) Get(bucket, key []byte) ([]byte, error)                  { return nil, nil }
func (nullPersistentState) Delete(bucket, key []byte) error                         { return nil }
func (nullPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error { return nil }
func (nullPersistentState) OpenOrCreate() error                                     { return nil }
func (nullPersistentState) Set(bucket, key, value []byte) error                     { return nil }

func newReadOnlyPersistentState(s PersistentState) PersistentState {
	return &readOnlyPersistentState{
		s: s,
	}
}

// Get implements PersistentState.Get.
func (s *readOnlyPersistentState) Get(bucket, key []byte) ([]byte, error) {
	return s.s.Get(bucket, key)
}

// Delete implements PersistentState.Delete.
func (s *readOnlyPersistentState) Delete(bucket, key []byte) error {
	return os.ErrPermission
}

// ForEach implements PersistentState.ForEach.
func (s *readOnlyPersistentState) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	return s.s.ForEach(bucket, fn)
}

// OpenOrCreate implements PersistentState.OpenOrCreate.
func (s *readOnlyPersistentState) OpenOrCreate() error {
	return s.s.OpenOrCreate()
}

// Set implements PersistentState.Set.
func (s *readOnlyPersistentState) Set(bucket, key, value []byte) error {
	return os.ErrPermission
}
