package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGetLongHelpPanics(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		mustLongHelp("non-existent-command")
	})
}
