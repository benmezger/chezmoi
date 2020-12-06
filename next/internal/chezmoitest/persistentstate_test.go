package chezmoitest

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistentState(t *testing.T) {
	t.Parallel()

	var (
		bucket = []byte("bucket")
		key    = []byte("key")
		value  = []byte("value")
	)

	s := NewPersistentState()

	require.NoError(t, s.OpenOrCreate())

	require.NoError(t, s.Delete(bucket, value))

	actualValue, err := s.Get(bucket, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)

	require.NoError(t, s.Set(bucket, key, value))

	actualValue, err = s.Get(bucket, key)
	require.NoError(t, err)
	assert.Equal(t, value, actualValue)

	require.NoError(t, s.ForEach(bucket, func(k, v []byte) error {
		assert.Equal(t, key, k)
		assert.Equal(t, value, v)
		return nil
	}))

	assert.Equal(t, io.EOF, s.ForEach(bucket, func(k, v []byte) error {
		return io.EOF
	}))

	require.NoError(t, s.Delete(bucket, key))
	actualValue, err = s.Get(bucket, key)
	require.NoError(t, err)
	assert.Nil(t, actualValue)
}
