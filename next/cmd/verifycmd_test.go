package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestVerifyCmd(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		root        interface{}
		expectedErr error
	}{
		{
			name: "empty",
			root: map[string]interface{}{
				"/home/user": &vfst.Dir{Perm: 0o700},
			},
		},
		{
			name: "file",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".bashrc":                         "# contents of .bashrc\n",
					".local/share/chezmoi/dot_bashrc": "# contents of .bashrc\n",
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			t.Cleanup(cleanup)

			assert.Equal(t, tc.expectedErr, newTestConfig(t, fs).execute([]string{"verify"}))
		})
	}
}
