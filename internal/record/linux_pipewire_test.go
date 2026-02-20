package record

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
		Duration:   200 * time.Millisecond,
	})
	require.NoError(t, err)

	argsRaw, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.NotContains(t, string(argsRaw), "--duration")
	require.Contains(t, string(argsRaw), "--rate")

	_, err = os.Stat(signalFile)
	require.NoError(t, err)
}

func TestPipeWireDurationModeReturnsContextCancellation(t *testing.T) {
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

	argsRaw, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(argsRaw), "--rate"))
}
