//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWritePIDFileCreatesFileWithCurrentPID(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "test.pid")
	err := writePIDFile(path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, strconv.Itoa(os.Getpid()), string(data))
}

func TestRemovePIDFileRemovesFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "test.pid")
	require.NoError(t, os.WriteFile(path, []byte("123"), 0o644))

	removePIDFile(path, zap.NewNop())

	_, err := os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestRemovePIDFileNonexistentIsNoop(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nonexistent.pid")
	removePIDFile(path, zap.NewNop())
}

func TestRegisterStopSignalClosesChannelOnSIGUSR1(t *testing.T) {
	stopCh := make(chan struct{})
	cleanup := registerStopSignal(stopCh, zap.NewNop())
	defer cleanup()

	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGUSR1))

	select {
	case <-stopCh:
	case <-time.After(2 * time.Second):
		t.Fatal("stopCh was not closed within timeout after SIGUSR1")
	}
}
