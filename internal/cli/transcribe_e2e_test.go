//go:build e2e

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

const (
	e2eWhisperPathEnv = "VOXCLIP_E2E_WHISPER_PATH"
	e2eModelDirEnv    = "VOXCLIP_E2E_MODEL_DIR"
)

func TestTranscribeEndToEndWithFSDD(t *testing.T) {
	whisperPath := strings.TrimSpace(os.Getenv(e2eWhisperPathEnv))
	if whisperPath == "" {
		t.Skip("set VOXCLIP_E2E_WHISPER_PATH to run e2e test")
	}

	modelDir := strings.TrimSpace(os.Getenv(e2eModelDirEnv))
	if modelDir == "" {
		modelDir = t.TempDir()
	}

	t.Setenv("VOXCLIP_WHISPER_PATH", whisperPath)

	_, setupStderr, err := runRootCommand(context.Background(), []string{
		"setup",
		"--model", "tiny",
		"--model-dir", modelDir,
		"--no-progress",
	})
	require.NoErrorf(t, err, "setup command failed: %s", setupStderr)

	fixtures := []struct {
		file            string
		expectedTokens  []string
		displayExpected string
	}{
		{file: "0_jackson_0.wav", expectedTokens: []string{"zero", "0"}, displayExpected: "zero"},
		{file: "1_jackson_0.wav", expectedTokens: []string{"one", "1"}, displayExpected: "one"},
		{file: "2_jackson_0.wav", expectedTokens: []string{"two", "2"}, displayExpected: "two"},
		{file: "3_jackson_0.wav", expectedTokens: []string{"three", "3"}, displayExpected: "three"},
		{file: "4_jackson_0.wav", expectedTokens: []string{"four", "4"}, displayExpected: "four"},
		{file: "5_jackson_0.wav", expectedTokens: []string{"five", "5"}, displayExpected: "five"},
		{file: "6_jackson_0.wav", expectedTokens: []string{"six", "6"}, displayExpected: "six"},
		{file: "7_jackson_0.wav", expectedTokens: []string{"seven", "7"}, displayExpected: "seven"},
		{file: "8_jackson_0.wav", expectedTokens: []string{"eight", "8"}, displayExpected: "eight"},
		{file: "9_jackson_0.wav", expectedTokens: []string{"nine", "9"}, displayExpected: "nine"},
	}

	matches := 0
	for _, fixture := range fixtures {
		audioPath := fsddFixturePath(t, fixture.file)

		stdout, stderr, err := runRootCommand(context.Background(), []string{
			"transcribe",
			"--model", "tiny",
			"--model-dir", modelDir,
			"--language", "en",
			"--no-progress",
			audioPath,
		})
		require.NoErrorf(t, err, "transcribe command failed for %s: %s", fixture.file, stderr)

		transcript := strings.TrimSpace(stdout)
		require.NotEmptyf(t, transcript, "empty transcript for %s", fixture.file)
		require.NotEqualf(t, blankAudioToken, transcript, "blank transcript for %s", fixture.file)

		normalized := normalizeTranscript(transcript)
		if containsAnyToken(normalized, fixture.expectedTokens) {
			matches++
			continue
		}

		t.Logf("fixture %s did not match expected token %q; transcript=%q normalized=%q", fixture.file, fixture.displayExpected, transcript, normalized)
	}

	require.GreaterOrEqual(t, matches, 7, "expected at least 7/10 fixtures to match expected spoken digits")
}

func fsddFixturePath(t *testing.T, fileName string) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "resolve current test file path")

	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	path := filepath.Join(root, "testdata", "audio", "fsdd", fileName)

	_, err := os.Stat(path)
	require.NoErrorf(t, err, "missing fixture %s", path)
	return path
}

func runRootCommand(ctx context.Context, args []string) (stdout string, stderr string, err error) {
	cmd := NewRootCmd()
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

func normalizeTranscript(input string) string {
	var b strings.Builder
	b.Grow(len(input))

	for _, r := range strings.ToLower(input) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteRune(' ')
		default:
			b.WriteRune(' ')
		}
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func containsAnyToken(normalized string, expected []string) bool {
	if normalized == "" {
		return false
	}

	fields := strings.Fields(normalized)
	set := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		set[field] = struct{}{}
	}

	for _, token := range expected {
		if _, ok := set[token]; ok {
			return true
		}
	}

	return false
}

func TestTranscribeBlankAudioEndToEnd(t *testing.T) {
	whisperPath := strings.TrimSpace(os.Getenv(e2eWhisperPathEnv))
	if whisperPath == "" {
		t.Skip("set VOXCLIP_E2E_WHISPER_PATH to run e2e test")
	}

	modelDir := strings.TrimSpace(os.Getenv(e2eModelDirEnv))
	if modelDir == "" {
		modelDir = t.TempDir()
	}

	t.Setenv("VOXCLIP_WHISPER_PATH", whisperPath)

	_, setupStderr, err := runRootCommand(context.Background(), []string{
		"setup",
		"--model", "tiny",
		"--model-dir", modelDir,
		"--no-progress",
	})
	require.NoErrorf(t, err, "setup command failed: %s", setupStderr)

	silentWAV := filepath.Join(t.TempDir(), "silent.wav")
	require.NoError(t, os.WriteFile(silentWAV, makePCM16WAVForTest(make([]int16, 16000), 16000, 1), 0o644))

	stdout, stderr, err := runRootCommand(context.Background(), []string{
		"transcribe",
		"--model", "tiny",
		"--model-dir", modelDir,
		"--no-progress",
		silentWAV,
	})
	require.NoErrorf(t, err, "transcribe command failed: %s", stderr)
	require.Equal(t, blankAudioToken, strings.TrimSpace(stdout))
}

func TestTranscribeSilenceGateBypassEndToEnd(t *testing.T) {
	whisperPath := strings.TrimSpace(os.Getenv(e2eWhisperPathEnv))
	if whisperPath == "" {
		t.Skip("set VOXCLIP_E2E_WHISPER_PATH to run e2e test")
	}

	modelDir := strings.TrimSpace(os.Getenv(e2eModelDirEnv))
	if modelDir == "" {
		modelDir = t.TempDir()
	}

	t.Setenv("VOXCLIP_WHISPER_PATH", whisperPath)

	_, setupStderr, err := runRootCommand(context.Background(), []string{
		"setup",
		"--model", "tiny",
		"--model-dir", modelDir,
		"--no-progress",
	})
	require.NoErrorf(t, err, "setup command failed: %s", setupStderr)

	silentWAV := filepath.Join(t.TempDir(), "silent.wav")
	require.NoError(t, os.WriteFile(silentWAV, makePCM16WAVForTest(make([]int16, 16000), 16000, 1), 0o644))

	_, stderr, err := runRootCommand(context.Background(), []string{
		"transcribe",
		"--model", "tiny",
		"--model-dir", modelDir,
		"--silence-gate=false",
		"--no-progress",
		silentWAV,
	})
	require.NoErrorf(t, err, "transcribe command failed: %s", stderr)
}

func TestTranscribeWithExplicitLanguageEndToEnd(t *testing.T) {
	whisperPath := strings.TrimSpace(os.Getenv(e2eWhisperPathEnv))
	if whisperPath == "" {
		t.Skip("set VOXCLIP_E2E_WHISPER_PATH to run e2e test")
	}

	modelDir := strings.TrimSpace(os.Getenv(e2eModelDirEnv))
	if modelDir == "" {
		modelDir = t.TempDir()
	}

	t.Setenv("VOXCLIP_WHISPER_PATH", whisperPath)

	_, setupStderr, err := runRootCommand(context.Background(), []string{
		"setup",
		"--model", "tiny",
		"--model-dir", modelDir,
		"--no-progress",
	})
	require.NoErrorf(t, err, "setup command failed: %s", setupStderr)

	audioPath := fsddFixturePath(t, "1_jackson_0.wav")

	tests := []struct {
		name     string
		language string
	}{
		{name: "explicit english", language: "en"},
		{name: "auto detect", language: "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runRootCommand(context.Background(), []string{
				"transcribe",
				"--model", "tiny",
				"--model-dir", modelDir,
				"--language", tt.language,
				"--no-progress",
				audioPath,
			})
			require.NoErrorf(t, err, "transcribe command failed: %s", stderr)

			transcript := strings.TrimSpace(stdout)
			require.NotEmptyf(t, transcript, "empty transcript with --language %s", tt.language)
			require.NotEqualf(t, blankAudioToken, transcript, "blank transcript with --language %s", tt.language)
		})
	}
}
