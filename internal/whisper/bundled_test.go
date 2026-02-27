package whisper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveBundledEnginePathFindsLibexecSibling(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	engineDir := filepath.Join(root, "libexec", "whisper")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	require.NoError(t, os.MkdirAll(engineDir, 0o755))

	voxclip := filepath.Join(binDir, "voxclip")
	require.NoError(t, os.WriteFile(voxclip, []byte(""), 0o755))

	enginePath := filepath.Join(engineDir, engineBinaryName())
	require.NoError(t, os.WriteFile(enginePath, []byte(""), 0o755))

	resolved, err := ResolveBundledEnginePath(voxclip)
	require.NoError(t, err)
	require.Equal(t, enginePath, resolved)
}

func TestResolveBundledEnginePathMissing(t *testing.T) {
	t.Parallel()

	voxclip := filepath.Join(t.TempDir(), "bin", "voxclip")
	require.NoError(t, os.MkdirAll(filepath.Dir(voxclip), 0o755))
	require.NoError(t, os.WriteFile(voxclip, []byte(""), 0o755))

	_, err := ResolveBundledEnginePath(voxclip)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bundled whisper engine not found")
}

func TestResolveBundledEnginePathFindsPackagingPathForLocalDev(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	voxclip := filepath.Join(root, "voxclip")
	require.NoError(t, os.WriteFile(voxclip, []byte(""), 0o755))

	targetDir := filepath.Join(root, "packaging", "whisper", fmt.Sprintf("%s_%s", runtime.GOOS, normalizeArch(runtime.GOARCH)))
	require.NoError(t, os.MkdirAll(targetDir, 0o755))
	enginePath := filepath.Join(targetDir, engineBinaryName())
	require.NoError(t, os.WriteFile(enginePath, []byte(""), 0o755))

	resolved, err := ResolveBundledEnginePath(voxclip)
	require.NoError(t, err)
	require.Equal(t, enginePath, resolved)
}

func TestIsMissingSharedLibraryError(t *testing.T) {
	t.Parallel()

	require.True(t, isMissingSharedLibraryError("error while loading shared libraries: libwhisper.so.1: cannot open shared object file"))
	require.True(t, isMissingSharedLibraryError("dyld: Library not loaded: @rpath/libwhisper.dylib"))
	require.False(t, isMissingSharedLibraryError("some other runtime error"))
}

func TestIsIllegalInstructionError(t *testing.T) {
	t.Parallel()

	require.True(t, isIllegalInstructionError("signal: illegal instruction (core dumped)"))
	require.True(t, isIllegalInstructionError("signal: illegal instruction"))
	require.False(t, isIllegalInstructionError("some other runtime error"))
	require.False(t, isIllegalInstructionError(""))
}
