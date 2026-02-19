package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranscribeCommandSkipsCopyForBlankTranscript(t *testing.T) {
	t.Parallel()

	out := new(bytes.Buffer)
	copyCalls := 0

	app := &appState{
		transcribeFn: func(_ context.Context, _ string) (string, error) {
			return "[BLANK_AUDIO]", nil
		},
		copyFn: func(_ context.Context, _ string) error {
			copyCalls++
			return nil
		},
	}

	cmd := newTranscribeCmd(app)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--copy", "/tmp/audio.wav"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Equal(t, 0, copyCalls)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
}

func TestTranscribeCommandCopiesBlankWhenCopyEmptyEnabled(t *testing.T) {
	t.Parallel()

	out := new(bytes.Buffer)
	copyCalls := 0

	app := &appState{
		copyEmpty: true,
		transcribeFn: func(_ context.Context, _ string) (string, error) {
			return "[BLANK_AUDIO]", nil
		},
		copyFn: func(_ context.Context, _ string) error {
			copyCalls++
			return nil
		},
	}

	cmd := newTranscribeCmd(app)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--copy", "/tmp/audio.wav"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Equal(t, 1, copyCalls)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
}
