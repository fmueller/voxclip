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
	require.NotNil(t, cmd.PersistentFlags().Lookup("model"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("model-dir"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("language"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("auto-download"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("backend"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("copy-empty"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("silence-gate"))
	require.NotNil(t, cmd.PersistentFlags().Lookup("silence-threshold-dbfs"))
	require.Equal(t, "true", cmd.PersistentFlags().Lookup("auto-download").DefValue)
	require.Equal(t, "false", cmd.PersistentFlags().Lookup("copy-empty").DefValue)
	require.Equal(t, "true", cmd.PersistentFlags().Lookup("silence-gate").DefValue)
	require.Equal(t, "-65", cmd.PersistentFlags().Lookup("silence-threshold-dbfs").DefValue)
	require.NotNil(t, cmd.Flags().Lookup("duration"))
	require.Equal(t, "0s", cmd.Flags().Lookup("duration").DefValue)
	require.NotNil(t, cmd.Flags().Lookup("immediate"))
	require.Equal(t, "false", cmd.Flags().Lookup("immediate").DefValue)
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
