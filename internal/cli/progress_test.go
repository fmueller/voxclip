package cli

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// safeBuffer is a goroutine-safe bytes.Buffer for capturing concurrent writes
// from progress bar goroutines during tests.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.buf.Bytes())
}

func TestStartSpinnerEnabled(t *testing.T) {
	t.Parallel()
	stop := startSpinner(io.Discard, true, "testing")
	require.NotNil(t, stop)
	stop()
}

func TestStartSpinnerDisabled(t *testing.T) {
	t.Parallel()
	stop := startSpinner(io.Discard, false, "testing")
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressEnabled(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(io.Discard, true, "testing", 5*time.Second)
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressDisabled(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(io.Discard, false, "testing", 5*time.Second)
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressZeroDuration(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(io.Discard, true, "testing", 0)
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressSubSecondDuration(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(io.Discard, true, "testing", 500*time.Millisecond)
	require.NotNil(t, stop)
	stop()
}

func TestSpinnerOutputEndsWithNewline(t *testing.T) {
	t.Parallel()
	var buf safeBuffer
	stop := startSpinner(&buf, true, "testing")
	time.Sleep(300 * time.Millisecond)
	stop()

	output := buf.Bytes()
	require.NotEmpty(t, output, "spinner should have written output")
	require.True(t, bytes.HasSuffix(output, []byte("\n")),
		"spinner output must end with newline to prevent log overlap, got trailing bytes: %q",
		trailingBytes(output, 20))
}

func TestDurationProgressOutputEndsWithNewline(t *testing.T) {
	t.Parallel()
	var buf safeBuffer
	stop := startDurationProgress(&buf, true, "testing", 5*time.Second)
	time.Sleep(300 * time.Millisecond)
	stop()

	output := buf.Bytes()
	require.NotEmpty(t, output, "duration progress should have written output")
	require.True(t, bytes.HasSuffix(output, []byte("\n")),
		"duration progress output must end with newline to prevent log overlap, got trailing bytes: %q",
		trailingBytes(output, 20))
}

func TestSpinnerDoubleStopIsSafe(t *testing.T) {
	t.Parallel()
	stop := startSpinner(io.Discard, true, "testing")
	time.Sleep(150 * time.Millisecond)
	stop()
	stop() // second call must not panic
}

func trailingBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[len(b)-n:]
}
