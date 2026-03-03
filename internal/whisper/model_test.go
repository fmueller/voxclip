package whisper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveModelDefaultNamedModel(t *testing.T) {
	t.Parallel()

	modelDir := t.TempDir()
	resolved, err := ResolveModel("", modelDir)
	require.NoError(t, err)

	wantName := DefaultModel()
	wantModel, ok := LookupModel(wantName)
	require.True(t, ok)

	require.Equal(t, wantName, resolved.Name)
	require.Equal(t, filepath.Join(modelDir, wantModel.FileName), resolved.Path)
	require.True(t, resolved.NeedsDownload)
	require.False(t, resolved.IsCustomPath)
}

func TestDefaultModelForLinux(t *testing.T) {
	t.Parallel()
	require.Equal(t, "tiny", DefaultModelForOS("linux"))
}

func TestDefaultModelForMacOS(t *testing.T) {
	t.Parallel()
	require.Equal(t, "small", DefaultModelForOS("darwin"))
}

func TestResolveModelExistingNamedModel(t *testing.T) {
	t.Parallel()

	modelDir := t.TempDir()
	modelPath := filepath.Join(modelDir, "ggml-tiny.bin")
	require.NoError(t, os.WriteFile(modelPath, []byte("ok"), 0o644))

	resolved, err := ResolveModel("tiny", modelDir)
	require.NoError(t, err)
	require.Equal(t, "tiny", resolved.Name)
	require.Equal(t, modelPath, resolved.Path)
	require.False(t, resolved.NeedsDownload)
}

func TestResolveModelCustomPath(t *testing.T) {
	t.Parallel()

	custom := filepath.Join(t.TempDir(), "custom.bin")
	require.NoError(t, os.WriteFile(custom, []byte("x"), 0o644))

	resolved, err := ResolveModel(custom, t.TempDir())
	require.NoError(t, err)
	require.True(t, resolved.IsCustomPath)
	require.Equal(t, custom, resolved.Path)
}

func TestResolveModelUnknownModel(t *testing.T) {
	t.Parallel()

	_, err := ResolveModel("super-huge", t.TempDir())
	require.Error(t, err)
}

func TestRegistryModelsHavePinnedChecksums(t *testing.T) {
	t.Parallel()

	for _, name := range ModelNames() {
		model, ok := LookupModel(name)
		require.True(t, ok)
		require.Lenf(t, model.SHA256, 64, "model %s should have pinned sha256", name)
	}
}
