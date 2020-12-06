package cmd

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
	xdg "github.com/twpayne/go-xdg/v3"

	"github.com/twpayne/chezmoi/next/internal/chezmoi"
	"github.com/twpayne/chezmoi/next/internal/chezmoitest"
)

func TestAddTemplateFuncPanic(t *testing.T) {
	t.Parallel()

	fs, cleanup, err := vfst.NewTestFS(nil)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	c := newTestConfig(t, fs)
	assert.NotPanics(t, func() {
		c.addTemplateFunc("func", nil)
	})
	assert.Panics(t, func() {
		c.addTemplateFunc("func", nil)
	})
}

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	t.Parallel()

	for s, expected := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		assert.Equal(t, expected, upperSnakeCaseToCamelCase(s))
	}
}

func TestValidateKeys(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		data        interface{}
		expectedErr bool
	}{
		{
			data:        nil,
			expectedErr: false,
		},
		{
			data: map[string]interface{}{
				"foo":                    "bar",
				"a":                      0,
				"_x9":                    false,
				"ThisVariableIsExported": nil,
				"αβ":                     "",
			},
			expectedErr: false,
		},
		{
			data: map[string]interface{}{
				"foo-foo": "bar",
			},
			expectedErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar-bar": "baz",
				},
			},
			expectedErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar-bar": "baz",
					},
				},
			},
			expectedErr: true,
		},
	} {
		if tc.expectedErr {
			assert.Error(t, validateKeys(tc.data, identifierRx))
		} else {
			assert.NoError(t, validateKeys(tc.data, identifierRx))
		}
	}
}

func newTestConfig(t *testing.T, fs vfs.FS) *Config {
	system := chezmoi.NewRealSystem(fs, chezmoitest.NewPersistentState())
	c, err := newConfig(
		withBaseSystem(system),
		withDestSystem(system),
		withSourceSystem(system),
		withTestFS(fs),
		withTestUser("user"),
	)
	require.NoError(t, err)
	return c
}

func withBaseSystem(baseSystem chezmoi.System) configOption {
	return func(c *Config) error {
		c.baseSystem = baseSystem
		return nil
	}
}

func withDestSystem(destSystem chezmoi.System) configOption {
	return func(c *Config) error {
		c.destSystem = destSystem
		return nil
	}
}

func withSourceSystem(sourceSystem chezmoi.System) configOption {
	return func(c *Config) error {
		c.sourceSystem = sourceSystem
		return nil
	}
}

func withTestFS(fs vfs.FS) configOption {
	return func(c *Config) error {
		c.fs = fs
		return nil
	}
}

func withTestUser(username string) configOption {
	return func(c *Config) error {
		var homeDirStr string
		switch runtime.GOOS {
		case "windows":
			homeDirStr = `C:\home\user`
		default:
			homeDirStr = "/home/user"
		}
		c.SourceDirStr = filepath.Join(homeDirStr, ".local", "share", "chezmoi")
		c.DestDirStr = homeDirStr
		c.Umask = 0o22
		configHome := filepath.Join(homeDirStr, ".config")
		dataHome := filepath.Join(homeDirStr, ".local", "share")
		c.bds = &xdg.BaseDirectorySpecification{
			ConfigHome: configHome,
			ConfigDirs: []string{configHome},
			DataHome:   dataHome,
			DataDirs:   []string{dataHome},
			CacheHome:  filepath.Join(homeDirStr, ".cache"),
			RuntimeDir: filepath.Join(homeDirStr, ".run"),
		}
		return nil
	}
}
