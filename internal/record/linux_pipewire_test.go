package record

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPipeWireDurationModeDoesNotUseDurationFlag(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")
	signalFile := filepath.Join(tempDir, "signal.txt")

	stubPath := filepath.Join(tempDir, "pw-record")
	stub := "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$@\" > \"$ARGS_FILE\"\ntrap 'touch \"$SIGNAL_FILE\"; exit 0' INT\nwhile :; do sleep 0.02; done\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	t.Setenv("PATH", tempDir+":"+os.Getenv("PATH"))
	t.Setenv("ARGS_FILE", argsFile)
	t.Setenv("SIGNAL_FILE", signalFile)

	backend := newPipeWireBackend()
	require.True(t, backend.Available())

	err := backend.Record(context.Background(), Config{
		OutputPath: filepath.Join(tempDir, "out.wav"),
		Duration:   500 * time.Millisecond,
	})
	require.NoError(t, err)

	waitForFile(t, argsFile, 500*time.Millisecond)
	argsRaw, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.NotContains(t, string(argsRaw), "--duration")
	require.Contains(t, string(argsRaw), "--rate")

	waitForFile(t, signalFile, 500*time.Millisecond)
}

func TestPipeWireDurationModeReturnsContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	stubPath := filepath.Join(tempDir, "pw-record")
	stub := "#!/bin/sh\nset -eu\nwhile :; do sleep 0.02; done\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	t.Setenv("PATH", tempDir+":"+os.Getenv("PATH"))

	backend := newPipeWireBackend()
	require.True(t, backend.Available())

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(80 * time.Millisecond)
		cancel()
	}()

	err := backend.Record(ctx, Config{
		OutputPath: filepath.Join(tempDir, "out.wav"),
		Duration:   3 * time.Second,
	})
	require.ErrorIs(t, err, context.Canceled)
}

func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		_, err := os.Stat(path)
		return err == nil
	}, timeout, 10*time.Millisecond)
}
