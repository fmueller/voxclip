package record

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubBackend struct {
	name      string
	available bool
}

func (s stubBackend) Name() string                                { return s.name }
func (s stubBackend) Available() bool                             { return s.available }
func (s stubBackend) Record(context.Context, Config) error        { return nil }
func (s stubBackend) ListDevices(context.Context) (string, error) { return "", nil }

func TestSelectBackendUsesPriorityOrder(t *testing.T) {
	t.Parallel()

	backend, err := SelectBackend([]Backend{
		stubBackend{name: "pw-record", available: false},
		stubBackend{name: "arecord", available: true},
		stubBackend{name: "ffmpeg", available: true},
	}, "auto")
	require.NoError(t, err)
	require.Equal(t, "arecord", backend.Name())
}

func TestSelectBackendUsesPreferredWhenAvailable(t *testing.T) {
	t.Parallel()

	backend, err := SelectBackend([]Backend{
		stubBackend{name: "pw-record", available: true},
		stubBackend{name: "arecord", available: true},
	}, "arecord")
	require.NoError(t, err)
	require.Equal(t, "arecord", backend.Name())
}

func TestSelectBackendReturnsErrorWhenUnavailable(t *testing.T) {
	t.Parallel()

	_, err := SelectBackend([]Backend{
		stubBackend{name: "pw-record", available: false},
	}, "pw-record")
	require.Error(t, err)
}

func TestSelectBackendReturnsErrorWhenNoBackendAvailable(t *testing.T) {
	t.Parallel()

	_, err := SelectBackend([]Backend{
		stubBackend{name: "pw-record", available: false},
		stubBackend{name: "arecord", available: false},
	}, "auto")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoBackendAvailable))
}
