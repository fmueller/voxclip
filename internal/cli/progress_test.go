package cli

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/fmueller/voxclip/internal/testutil"
	"github.com/stretchr/testify/require"
)

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

func TestSpinnerOutputEndsWithCarriageReturn(t *testing.T) {
	t.Parallel()
	var buf testutil.SafeBuffer
	stop := startSpinner(&buf, true, "testing")
	time.Sleep(300 * time.Millisecond)
	stop()

	output := buf.Bytes()
	require.NotEmpty(t, output, "spinner should have written output")
	require.True(t, bytes.HasSuffix(output, []byte("\r")),
		"spinner output must end with carriage return to clear the line, got trailing bytes: %q",
		testutil.TrailingBytes(output, 20))
}

func TestDurationProgressOutputEndsWithNewline(t *testing.T) {
	t.Parallel()
	var buf testutil.SafeBuffer
	stop := startDurationProgress(&buf, true, "testing", 5*time.Second)
	time.Sleep(300 * time.Millisecond)
	stop()

	output := buf.Bytes()
	require.NotEmpty(t, output, "duration progress should have written output")
	require.True(t, bytes.HasSuffix(output, []byte("\n")),
		"duration progress output must end with newline to prevent log overlap, got trailing bytes: %q",
		testutil.TrailingBytes(output, 20))
}

func TestSpinnerDoubleStopIsSafe(t *testing.T) {
	t.Parallel()
	stop := startSpinner(io.Discard, true, "testing")
	time.Sleep(150 * time.Millisecond)
	stop()
	stop() // second call must not panic
}
