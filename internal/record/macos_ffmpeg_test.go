package record

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFFMPEGMacListDevicesReturnsOutputOnNonZeroExit(t *testing.T) {
	tempDir := t.TempDir()

	stubPath := filepath.Join(tempDir, "ffmpeg")
	stub := `#!/bin/sh
>&2 echo "[AVFoundation indev] AVFoundation audio devices:"
>&2 echo "[AVFoundation indev] [0] Built-in Microphone"
exit 1
`
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	t.Setenv("PATH", tempDir+":"+os.Getenv("PATH"))

	backend := &ffmpegMacBackend{}
	require.True(t, backend.Available())

	out, err := backend.ListDevices(context.Background())
	require.NoError(t, err)
	require.Contains(t, out, "Built-in Microphone")
}
