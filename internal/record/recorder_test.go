package record

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubBackend struct {
	name      string
	available bool
	recordErr error
	recordFn  func(Config) error
	calls     int
}

func (s *stubBackend) Name() string    { return s.name }
func (s *stubBackend) Available() bool { return s.available }
func (s *stubBackend) Record(_ context.Context, cfg Config) error {
	s.calls++
	if s.recordFn != nil {
		return s.recordFn(cfg)
	}
	return s.recordErr
}
func (s *stubBackend) ListDevices(context.Context) (string, error) { return "", nil }

func TestSelectBackendUsesPriorityOrder(t *testing.T) {
	t.Parallel()

	backend, err := SelectBackend([]Backend{
		&stubBackend{name: "pw-record", available: false},
		&stubBackend{name: "arecord", available: true},
		&stubBackend{name: "ffmpeg", available: true},
	}, "auto")
	require.NoError(t, err)
	require.Equal(t, "arecord", backend.Name())
}

func TestSelectBackendUsesPreferredWhenAvailable(t *testing.T) {
	t.Parallel()

	backend, err := SelectBackend([]Backend{
		&stubBackend{name: "pw-record", available: true},
		&stubBackend{name: "arecord", available: true},
	}, "arecord")
	require.NoError(t, err)
	require.Equal(t, "arecord", backend.Name())
}

func TestSelectBackendReturnsErrorWhenUnavailable(t *testing.T) {
	t.Parallel()

	_, err := SelectBackend([]Backend{
		&stubBackend{name: "pw-record", available: false},
	}, "pw-record")
	require.Error(t, err)
}

func TestSelectBackendReturnsErrorWhenNoBackendAvailable(t *testing.T) {
	t.Parallel()

	_, err := SelectBackend([]Backend{
		&stubBackend{name: "pw-record", available: false},
		&stubBackend{name: "arecord", available: false},
	}, "auto")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoBackendAvailable))
}

func TestRecordWithFallbackUsesNextBackendAfterFailure(t *testing.T) {
	t.Parallel()

	first := &stubBackend{name: "pw-record", available: true, recordErr: errors.New("boom")}
	second := &stubBackend{name: "arecord", available: true}

	usedBackend, err := recordWithFallback(context.Background(), []Backend{first, second}, "auto", Config{OutputPath: filepath.Join(t.TempDir(), "audio.wav")})
	require.NoError(t, err)
	require.Equal(t, "arecord", usedBackend)
	require.Equal(t, 1, first.calls)
	require.Equal(t, 1, second.calls)
}

func TestRecordWithFallbackUsesPreferredBackendFirst(t *testing.T) {
	t.Parallel()

	pw := &stubBackend{name: "pw-record", available: true}
	alsa := &stubBackend{name: "arecord", available: true, recordErr: errors.New("failed")}
	ffmpeg := &stubBackend{name: "ffmpeg", available: true}

	usedBackend, err := recordWithFallback(context.Background(), []Backend{pw, alsa, ffmpeg}, "arecord", Config{OutputPath: filepath.Join(t.TempDir(), "audio.wav")})
	require.NoError(t, err)
	require.Equal(t, "pw-record", usedBackend)
	require.Equal(t, 1, alsa.calls)
	require.Equal(t, 1, pw.calls)
	require.Equal(t, 0, ffmpeg.calls)
}

func TestRecordWithFallbackSkipsUnavailableBackends(t *testing.T) {
	t.Parallel()

	unavailable := &stubBackend{name: "pw-record", available: false}
	working := &stubBackend{name: "arecord", available: true}

	usedBackend, err := recordWithFallback(context.Background(), []Backend{unavailable, working}, "auto", Config{OutputPath: filepath.Join(t.TempDir(), "audio.wav")})
	require.NoError(t, err)
	require.Equal(t, "arecord", usedBackend)
	require.Equal(t, 0, unavailable.calls)
	require.Equal(t, 1, working.calls)
}

func TestRecordWithFallbackReturnsErrorWhenAllBackendsFail(t *testing.T) {
	t.Parallel()

	pw := &stubBackend{name: "pw-record", available: true, recordErr: errors.New("pw failed")}
	alsa := &stubBackend{name: "arecord", available: true, recordErr: errors.New("alsa failed")}

	_, err := recordWithFallback(context.Background(), []Backend{pw, alsa}, "auto", Config{OutputPath: filepath.Join(t.TempDir(), "audio.wav")})
	require.Error(t, err)
	require.ErrorContains(t, err, "record audio with available backends")
	require.ErrorContains(t, err, "pw-record")
	require.ErrorContains(t, err, "arecord")
}

func TestRecordWithFallbackRemovesPartialOutputBeforeRetry(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "audio.wav")
	first := &stubBackend{
		name:      "pw-record",
		available: true,
		recordFn: func(cfg Config) error {
			if err := os.WriteFile(cfg.OutputPath, []byte("partial"), 0o644); err != nil {
				return err
			}
			return errors.New("failed")
		},
	}
	second := &stubBackend{
		name:      "arecord",
		available: true,
		recordFn: func(cfg Config) error {
			_, err := os.Stat(cfg.OutputPath)
			if err == nil {
				return errors.New("partial output still exists")
			}
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			return nil
		},
	}

	usedBackend, err := recordWithFallback(context.Background(), []Backend{first, second}, "auto", Config{OutputPath: outputPath})
	require.NoError(t, err)
	require.Equal(t, "arecord", usedBackend)
}
