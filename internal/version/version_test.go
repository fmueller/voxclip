package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve_ReturnsVersion(t *testing.T) {
	original := Version
	Version = "1.2.3-test"
	t.Cleanup(func() { Version = original })
	require.Equal(t, "1.2.3-test", Resolve())
}

func TestResolve_EmptyFallsBackToZero(t *testing.T) {
	original := Version
	Version = ""
	t.Cleanup(func() { Version = original })
	require.Equal(t, "0.0.0", Resolve())
}
