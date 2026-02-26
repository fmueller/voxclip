package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		errContains string
	}{
		{
			name:        "unknown command",
			args:        []string{"badcmd"},
			errContains: "unknown command",
		},
		{
			name:        "unknown root flag",
			args:        []string{"--badflag"},
			errContains: "unknown flag",
		},
		{
			name:        "unknown subcommand flag",
			args:        []string{"transcribe", "--bogus", "f.wav"},
			errContains: "unknown flag",
		},
		{
			name:        "transcribe missing arg",
			args:        []string{"transcribe"},
			errContains: "accepts 1 arg(s)",
		},
		{
			name:        "transcribe too many args",
			args:        []string{"transcribe", "a.wav", "b.wav"},
			errContains: "accepts 1 arg(s)",
		},
		{
			name:        "transcribe nonexistent file",
			args:        []string{"transcribe", "/no/such/file.wav"},
			errContains: "audio file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := runCommand(t, tt.args)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestSetupRejectsNonexistentCustomModelPath(t *testing.T) {
	t.Parallel()

	_, _, err := runCommand(t, []string{"setup", "--model", "/no/such/path/model.bin"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "custom model path does not exist")
}

func TestVersionFlagOutput(t *testing.T) {
	t.Parallel()

	stdout, _, err := runCommand(t, []string{"--version"})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(stdout, "voxclip v"), "expected version prefix, got: %s", stdout)
}
