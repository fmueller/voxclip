package cli

import "strings"

const blankAudioToken = "[BLANK_AUDIO]"

func isBlankTranscript(transcript string) bool {
	trimmed := strings.TrimSpace(transcript)
	if trimmed == "" {
		return true
	}

	return strings.EqualFold(trimmed, blankAudioToken)
}

func noSpeechHint() string {
	return "No speech detected. Check mic mute and selected input device, then try again."
}
