package chezmoi

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest,tparallel
func TestOSPathFormat(t *testing.T) {
	t.Parallel()

	type s struct {
		Dir *OSPath
	}

	for name, format := range Formats {
		t.Run(name, func(t *testing.T) {
			var dirStr string
			switch runtime.GOOS {
			case "windows":
				dirStr = `C:\home\user`
			default:
				dirStr = "/home/user"
			}
			expectedS := &s{
				Dir: NewOSPath(dirStr),
			}
			data, err := format.Marshal(expectedS)
			assert.NoError(t, err)
			actualS := &s{}
			assert.NoError(t, format.Decode(data, actualS))
			assert.Equal(t, expectedS, actualS)
		})
	}
}
