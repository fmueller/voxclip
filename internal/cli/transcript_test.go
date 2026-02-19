package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsBlankTranscript(t *testing.T) {
	t.Parallel()

	require.True(t, isBlankTranscript(""))
	require.True(t, isBlankTranscript("   \n\t "))
	require.True(t, isBlankTranscript("[BLANK_AUDIO]"))
	require.True(t, isBlankTranscript(" [blank_audio] "))
	require.False(t, isBlankTranscript("Hello world"))
}
