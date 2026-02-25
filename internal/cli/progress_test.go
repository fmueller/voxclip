package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStartSpinnerEnabled(t *testing.T) {
	t.Parallel()
	stop := startSpinner(true, "testing")
	require.NotNil(t, stop)
	stop()
}

func TestStartSpinnerDisabled(t *testing.T) {
	t.Parallel()
	stop := startSpinner(false, "testing")
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressEnabled(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(true, "testing", 5*time.Second)
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressDisabled(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(false, "testing", 5*time.Second)
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressZeroDuration(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(true, "testing", 0)
	require.NotNil(t, stop)
	stop()
}

func TestStartDurationProgressSubSecondDuration(t *testing.T) {
	t.Parallel()
	stop := startDurationProgress(true, "testing", 500*time.Millisecond)
	require.NotNil(t, stop)
	stop()
}
