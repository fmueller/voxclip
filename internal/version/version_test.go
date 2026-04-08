package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolve_ReturnsVersion(t *testing.T) {
	t.Parallel()
	require.Equal(t, Version, Resolve())
}

func TestResolve_EmptyFallsBackToZero(t *testing.T) {
	t.Parallel()
	original := Version
	Version = ""
	t.Cleanup(func() { Version = original })
	require.Equal(t, "0.0.0", Resolve())
}
