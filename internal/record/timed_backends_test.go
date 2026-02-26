package record

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestALSADurationModeReturnsContextCancellation(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "arecord", false)

	backend := newALSARecorderBackend()
	require.True(t, backend.Available())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(ctx, Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			Duration:   3 * time.Second,
		})
	}()
	t.Cleanup(cancel)

	waitForFile(t, readyFile, 5*time.Second)
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestFFMPEGLinuxDurationModeReturnsContextCancellation(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "ffmpeg", false)

	backend := newFFMPEGLinuxBackend()
	require.True(t, backend.Available())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(ctx, Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			Duration:   3 * time.Second,
			Format:     "pulse",
		})
	}()
	t.Cleanup(cancel)

	waitForFile(t, readyFile, 5*time.Second)
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestFFMPEGMacDurationModeReturnsContextCancellation(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "ffmpeg", false)

	backend := newFFMPEGMacOSBackend()
	require.True(t, backend.Available())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(ctx, Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			Duration:   3 * time.Second,
		})
	}()
	t.Cleanup(cancel)

	waitForFile(t, readyFile, 5*time.Second)
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestRunTimedCommandKillsWhenInterruptIgnored(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "ignore-int", true)

	cmd := exec.Command(filepath.Join(tempDir, "ignore-int"))
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runTimedCommand(ctx, cmd, 3*time.Second, nil)
	}()
	t.Cleanup(cancel)

	waitForFile(t, readyFile, 5*time.Second)
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestRunTimedCommandKillsOnTimerWhenInterruptIgnored(t *testing.T) {
	tempDir, _ := setupRunningCommandStub(t, "ignore-int", true)

	cmd := exec.Command(filepath.Join(tempDir, "ignore-int"))
	errCh := make(chan error, 1)
	go func() {
		errCh <- runTimedCommand(context.Background(), cmd, 100*time.Millisecond, nil)
	}()

	err := <-errCh
	require.NoError(t, err)
}

func setupRunningCommandStub(t *testing.T, name string, ignoreInterrupt bool) (string, string) {
	t.Helper()

	tempDir := t.TempDir()
	readyFile := filepath.Join(tempDir, "ready.txt")

	trap := "trap 'exit 0' INT"
	if ignoreInterrupt {
		trap = "trap '' INT"
	}

	stubPath := filepath.Join(tempDir, name)
	stub := "#!/bin/sh\nset -eu\ntouch \"$READY_FILE\"\n" + trap + "\nwhile :; do sleep 0.02; done\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	t.Setenv("PATH", tempDir+":"+os.Getenv("PATH"))
	t.Setenv("READY_FILE", readyFile)

	return tempDir, readyFile
}
