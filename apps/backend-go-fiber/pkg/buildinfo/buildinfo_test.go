package buildinfo

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaults(t *testing.T) {
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "unknown", BuildTime)
}

func TestGoVersion(t *testing.T) {
	assert.Equal(t, runtime.Version(), GoVersion())
}
