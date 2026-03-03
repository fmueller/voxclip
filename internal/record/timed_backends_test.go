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

func TestRunSignalStopCommandStopsOnChannel(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "signal-stop", false)

	cmd := exec.Command(filepath.Join(tempDir, "signal-stop"))
	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(context.Background(), cmd, stopCh, 0, nil)
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestRunSignalStopCommandReturnsContextCancellation(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "signal-stop-ctx", false)

	cmd := exec.Command(filepath.Join(tempDir, "signal-stop-ctx"))
	stopCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(ctx, cmd, stopCh, 0, nil)
	}()
	t.Cleanup(cancel)

	waitForFile(t, readyFile, 5*time.Second)
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestRunSignalStopCommandKillsWhenInterruptIgnored(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "signal-stop-ign", true)

	cmd := exec.Command(filepath.Join(tempDir, "signal-stop-ign"))
	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(context.Background(), cmd, stopCh, 0, nil)
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestRunSignalStopCommandSubprocessExitsOnItsOwn(t *testing.T) {
	tempDir := t.TempDir()
	stubPath := filepath.Join(tempDir, "quick-exit")
	stub := "#!/bin/sh\nexit 0\n"
	require.NoError(t, os.WriteFile(stubPath, []byte(stub), 0o755))

	cmd := exec.Command(stubPath)
	stopCh := make(chan struct{})
	err := runSignalStopCommand(context.Background(), cmd, stopCh, 0, nil)
	require.NoError(t, err)
}

func TestPipeWireSignalStopModeStopsOnChannel(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "pw-record", false)

	backend := newPipeWireBackend()
	require.True(t, backend.Available())

	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(context.Background(), Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			StopCh:     stopCh,
		})
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestALSASignalStopModeStopsOnChannel(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "arecord", false)

	backend := newALSARecorderBackend()
	require.True(t, backend.Available())

	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(context.Background(), Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			StopCh:     stopCh,
		})
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestFFMPEGLinuxSignalStopModeStopsOnChannel(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "ffmpeg", false)

	backend := newFFMPEGLinuxBackend()
	require.True(t, backend.Available())

	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(context.Background(), Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			Format:     "pulse",
			StopCh:     stopCh,
		})
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestFFMPEGMacSignalStopModeStopsOnChannel(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "ffmpeg", false)

	backend := newFFMPEGMacOSBackend()
	require.True(t, backend.Available())

	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- backend.Record(context.Background(), Config{
			OutputPath: filepath.Join(tempDir, "out.wav"),
			StopCh:     stopCh,
		})
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

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

func TestRunSignalStopCommandStopsOnMaxDuration(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "duration-stop", false)

	cmd := exec.Command(filepath.Join(tempDir, "duration-stop"))
	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(context.Background(), cmd, stopCh, 200*time.Millisecond, nil)
	}()

	waitForFile(t, readyFile, 5*time.Second)

	err := <-errCh
	require.NoError(t, err)
}

func TestRunSignalStopCommandSignalBeforeMaxDuration(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "signal-first", false)

	cmd := exec.Command(filepath.Join(tempDir, "signal-first"))
	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(context.Background(), cmd, stopCh, 10*time.Second, nil)
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestRunSignalStopCommandZeroMaxDurationWaitsForSignal(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "zero-dur", false)

	cmd := exec.Command(filepath.Join(tempDir, "zero-dur"))
	stopCh := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(context.Background(), cmd, stopCh, 0, nil)
	}()

	waitForFile(t, readyFile, 5*time.Second)
	close(stopCh)

	err := <-errCh
	require.NoError(t, err)
}

func TestRunSignalStopCommandContextCancelWithMaxDuration(t *testing.T) {
	tempDir, readyFile := setupRunningCommandStub(t, "ctx-dur", false)

	cmd := exec.Command(filepath.Join(tempDir, "ctx-dur"))
	stopCh := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSignalStopCommand(ctx, cmd, stopCh, 10*time.Second, nil)
	}()
	t.Cleanup(cancel)

	waitForFile(t, readyFile, 5*time.Second)
	cancel()

	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}
