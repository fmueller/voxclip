//go:build integration

package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func noopPreflight(_ context.Context) error { return nil }

func TestRunDefaultFlowSuccess(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)
	audioFile := filepath.Join(t.TempDir(), "audio.wav")
	require.NoError(t, os.WriteFile(audioFile, []byte("fake"), 0o644))

	app := &appState{
		out:         out,
		preflightFn: noopPreflight,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return audioFile, nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "hello world", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "hello world\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:" + audioFile,
		"copy:hello world",
	}, order)
	_, statErr := os.Stat(audioFile)
	require.ErrorIs(t, statErr, os.ErrNotExist, "recording should be removed after default flow")
}

func TestRunDefaultClipboardFailureIsNonFatal(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)
	audioFile := filepath.Join(t.TempDir(), "audio.wav")
	require.NoError(t, os.WriteFile(audioFile, []byte("fake"), 0o644))

	app := &appState{
		out:         out,
		preflightFn: noopPreflight,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return audioFile, nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "clipboard fallback", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return errors.New("clipboard command failed")
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "clipboard fallback\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:" + audioFile,
		"copy:clipboard fallback",
	}, order)
	_, statErr := os.Stat(audioFile)
	require.ErrorIs(t, statErr, os.ErrNotExist, "recording should be removed after default flow")
}

func TestRunDefaultSkipsCopyForBlankTranscript(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)
	audioFile := filepath.Join(t.TempDir(), "audio.wav")
	require.NoError(t, os.WriteFile(audioFile, []byte("fake"), 0o644))

	app := &appState{
		out:         out,
		preflightFn: noopPreflight,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return audioFile, nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "[BLANK_AUDIO]", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:" + audioFile,
	}, order)
	_, statErr := os.Stat(audioFile)
	require.ErrorIs(t, statErr, os.ErrNotExist, "recording should be removed after default flow")
}

func TestRunDefaultCopiesBlankWhenCopyEmptyEnabled(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)
	audioFile := filepath.Join(t.TempDir(), "audio.wav")
	require.NoError(t, os.WriteFile(audioFile, []byte("fake"), 0o644))

	app := &appState{
		out:         out,
		copyEmpty:   true,
		preflightFn: noopPreflight,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return audioFile, nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "[BLANK_AUDIO]", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:" + audioFile,
		"copy:[BLANK_AUDIO]",
	}, order)
	_, statErr := os.Stat(audioFile)
	require.ErrorIs(t, statErr, os.ErrNotExist, "recording should be removed after default flow")
}

func TestRunDefaultSkipsTranscribeWhenRecordingIsSilent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "silent.wav")
	require.NoError(t, os.WriteFile(path, makePCM16WAVForTest(make([]int16, 16000), 16000, 1), 0o644))

	out := new(bytes.Buffer)
	transcribeCalls := 0
	copyCalls := 0

	app := &appState{
		out:         out,
		silenceGate: true,
		silenceDBFS: -65,
		preflightFn: noopPreflight,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			return path, nil
		},
		transcribeFn: func(_ context.Context, _ string) (string, error) {
			transcribeCalls++
			return "should-not-happen", nil
		},
		copyFn: func(_ context.Context, _ string) error {
			copyCalls++
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, transcribeCalls)
	require.Equal(t, 0, copyCalls)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
	_, statErr := os.Stat(path)
	require.ErrorIs(t, statErr, os.ErrNotExist, "recording should be removed after default flow")
}

func TestRunDefaultForwardsDurationToRecordOptions(t *testing.T) {
	out := new(bytes.Buffer)
	audioFile := filepath.Join(t.TempDir(), "audio.wav")
	require.NoError(t, os.WriteFile(audioFile, []byte("fake"), 0o644))

	var captured recordOptions
	app := &appState{
		out:         out,
		duration:    5 * time.Second,
		preflightFn: noopPreflight,
		recordFn: func(_ context.Context, opts recordOptions) (string, error) {
			captured = opts
			return audioFile, nil
		},
		transcribeFn: func(_ context.Context, _ string) (string, error) {
			return "hello", nil
		},
		copyFn: func(_ context.Context, _ string) error {
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, 5*time.Second, captured.duration)
}

func TestRunDefaultPreflightErrorAbortsBeforeRecording(t *testing.T) {
	recorded := false
	app := &appState{
		model:        "nonexistent-model",
		modelDir:     t.TempDir(),
		autoDownload: false,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			recorded = true
			return "/tmp/audio.wav", nil
		},
		transcribeFn: func(_ context.Context, _ string) (string, error) {
			return "hello", nil
		},
		copyFn: func(_ context.Context, _ string) error {
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.Error(t, err)
	require.False(t, recorded, "recording should not happen when preflight fails")
}
