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

func TestSanitizeLanguage(t *testing.T) {
	t.Parallel()

	require.Equal(t, "auto", sanitizeLanguage(""))
	require.Equal(t, "auto", sanitizeLanguage("   "))
	require.Equal(t, "en", sanitizeLanguage("en"))
	require.Equal(t, "en", sanitizeLanguage(" EN "))
	require.Equal(t, "de", sanitizeLanguage("De"))
}
