package record

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestArecordInputPassesDevice(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")

	stubPath := filepath.Join(tempDir, "arecord")
	stub := "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$@\" > \"$ARGS_FILE\"\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	t.Setenv("PATH", tempDir+":"+os.Getenv("PATH"))
	t.Setenv("ARGS_FILE", argsFile)

	backend := newALSARecorderBackend()
	require.True(t, backend.Available())

	err := backend.Record(context.Background(), Config{
		OutputPath: filepath.Join(tempDir, "out.wav"),
		Input:      "hw:1,0",
	})
	require.NoError(t, err)

	argsRaw, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	args := string(argsRaw)
	require.Contains(t, args, "-D\nhw:1,0\n")
}

func TestArecordNoInputOmitsDevice(t *testing.T) {
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "args.txt")

	stubPath := filepath.Join(tempDir, "arecord")
	stub := "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$@\" > \"$ARGS_FILE\"\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	t.Setenv("PATH", tempDir+":"+os.Getenv("PATH"))
	t.Setenv("ARGS_FILE", argsFile)

	backend := newALSARecorderBackend()
	require.True(t, backend.Available())

	err := backend.Record(context.Background(), Config{
		OutputPath: filepath.Join(tempDir, "out.wav"),
	})
	require.NoError(t, err)

	argsRaw, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.NotContains(t, string(argsRaw), "-D")
}
