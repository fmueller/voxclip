package platform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultModelDirForLinuxWithXDG(t *testing.T) {
	t.Parallel()

	dir, err := DefaultModelDirFor("linux", "/home/dev", "/tmp/xdg-data")
	require.NoError(t, err)
	require.Equal(t, "/tmp/xdg-data/voxclip/models", dir)
}

func TestDefaultModelDirForLinuxWithoutXDG(t *testing.T) {
	t.Parallel()

	dir, err := DefaultModelDirFor("linux", "/home/dev", "")
	require.NoError(t, err)
	require.Equal(t, "/home/dev/.local/share/voxclip/models", dir)
}

func TestDefaultModelDirForMacOS(t *testing.T) {
	t.Parallel()

	dir, err := DefaultModelDirFor("darwin", "/Users/dev", "")
	require.NoError(t, err)
	require.Equal(t, "/Users/dev/Library/Application Support/voxclip/models", dir)
}

func TestDefaultModelDirForUnsupportedOS(t *testing.T) {
	t.Parallel()

	_, err := DefaultModelDirFor("windows", "/Users/dev", "")
	require.Error(t, err)
}
