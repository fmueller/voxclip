package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommandRegistersCoreSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewRootCmd()

	require.NotNil(t, cmd.Commands())
	require.NotNil(t, cmd.Flags().Lookup("verbose"))
	require.NotNil(t, cmd.Flags().Lookup("json"))
	require.NotNil(t, cmd.Flags().Lookup("no-progress"))
	require.NotNil(t, cmd.Flags().Lookup("model"))
	require.NotNil(t, cmd.Flags().Lookup("model-dir"))
	require.NotNil(t, cmd.Flags().Lookup("language"))
	require.NotNil(t, cmd.Flags().Lookup("auto-download"))
	require.NotNil(t, cmd.Flags().Lookup("backend"))
	require.NotNil(t, cmd.Flags().Lookup("input"))
	require.NotNil(t, cmd.Flags().Lookup("input-format"))
	require.NotNil(t, cmd.Flags().Lookup("copy-empty"))
	require.NotNil(t, cmd.Flags().Lookup("silence-gate"))
	require.NotNil(t, cmd.Flags().Lookup("silence-threshold-dbfs"))
	require.Equal(t, "true", cmd.Flags().Lookup("auto-download").DefValue)
	require.Equal(t, "false", cmd.Flags().Lookup("copy-empty").DefValue)
	require.Equal(t, "true", cmd.Flags().Lookup("silence-gate").DefValue)
	require.Equal(t, "-65", cmd.Flags().Lookup("silence-threshold-dbfs").DefValue)
	require.NotNil(t, cmd.Flags().Lookup("duration"))
	require.Equal(t, "0s", cmd.Flags().Lookup("duration").DefValue)
	require.NotNil(t, cmd.Flags().Lookup("immediate"))
	require.Equal(t, "false", cmd.Flags().Lookup("immediate").DefValue)
	require.Nil(t, cmd.PersistentFlags().Lookup("model"))
}

func TestRootHelpParsesSuccessfully(t *testing.T) {
	t.Parallel()

	cmd := NewRootCmd()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "record")
	require.Contains(t, out.String(), "transcribe")
	require.Contains(t, out.String(), "setup")
	require.Contains(t, out.String(), "devices")
	require.Contains(t, out.String(), "version")
}

func TestSubcommandHelpParsesSuccessfully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{name: "record", args: []string{"record", "--help"}, contains: "Record audio into a WAV file"},
		{name: "transcribe", args: []string{"transcribe", "--help"}, contains: "Transcribe an audio file"},
		{name: "devices", args: []string{"devices", "--help"}, contains: "List recording devices"},
		{name: "setup", args: []string{"setup", "--help"}, contains: "Download and verify speech model assets"},
		{name: "version", args: []string{"version", "--help"}, contains: "Print the version number"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewRootCmd()
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			require.NoError(t, err)
			require.Contains(t, out.String(), tt.contains)
		})
	}
}

func TestSubcommandsRejectIrrelevantFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "record rejects model", args: []string{"record", "--model", "small"}},
		{name: "transcribe rejects backend", args: []string{"transcribe", "--backend", "auto", "/tmp/audio.wav"}},
		{name: "setup rejects language", args: []string{"setup", "--language", "de"}},
		{name: "devices rejects verbose", args: []string{"devices", "--verbose"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewRootCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			require.Error(t, err)
			require.Contains(t, err.Error(), "unknown flag")
		})
	}
}

func TestHelpShowsStrictFlagScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		contains    []string
		notContains []string
	}{
		{
			name: "root help has default flow flags",
			args: []string{"--help"},
			contains: []string{
				"--model string",
				"--backend string",
				"--copy-empty",
				"--duration duration",
				"--input string",
			},
		},
		{
			name: "record help omits transcription flags",
			args: []string{"record", "--help"},
			contains: []string{
				"--backend string",
				"--duration duration",
			},
			notContains: []string{
				"--model string",
				"--language string",
				"Global Flags:",
			},
		},
		{
			name: "transcribe help omits recording flags",
			args: []string{"transcribe", "--help"},
			contains: []string{
				"--model string",
				"--copy",
			},
			notContains: []string{
				"--backend string",
				"--input-format",
				"Global Flags:",
			},
		},
		{
			name: "setup help omits recording runtime flags",
			args: []string{"setup", "--help"},
			contains: []string{
				"--model string",
				"--model-dir string",
			},
			notContains: []string{
				"--backend string",
				"--language string",
				"Global Flags:",
			},
		},
		{
			name: "devices help has no operational flags",
			args: []string{"devices", "--help"},
			notContains: []string{
				"--verbose",
				"--model string",
				"Global Flags:",
			},
		},
		{
			name: "version help has no operational flags",
			args: []string{"version", "--help"},
			notContains: []string{
				"--verbose",
				"--model string",
				"Global Flags:",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewRootCmd()
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			require.NoError(t, err)

			output := out.String()
			for _, item := range tt.contains {
				require.Contains(t, output, item)
			}
			for _, item := range tt.notContains {
				require.NotContains(t, output, item)
			}
		})
	}
}

func TestVersionCommandOutputMatchesFlag(t *testing.T) {
	t.Parallel()

	flagOut := new(bytes.Buffer)
	flagCmd := NewRootCmd()
	flagCmd.SetOut(flagOut)
	flagCmd.SetErr(flagOut)
	flagCmd.SetArgs([]string{"--version"})
	require.NoError(t, flagCmd.Execute())

	subOut := new(bytes.Buffer)
	subCmd := NewRootCmd()
	subCmd.SetOut(subOut)
	subCmd.SetErr(subOut)
	subCmd.SetArgs([]string{"version"})
	require.NoError(t, subCmd.Execute())

	require.Equal(t, flagOut.String(), subOut.String())
}
