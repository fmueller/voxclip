package main

import (
	"errors"
	"testing"

	"github.com/fmueller/voxclip/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestShouldPrintUsageHint(t *testing.T) {
	t.Parallel()

	require.True(t, shouldPrintUsageHint(errors.New("unknown command \"bad\" for \"voxclip\"")))
	require.True(t, shouldPrintUsageHint(errors.New("unknown flag: --oops")))
	require.True(t, shouldPrintUsageHint(errors.New("accepts 1 arg(s), received 0")))
	require.False(t, shouldPrintUsageHint(errors.New("download model \"small\": context deadline exceeded")))
	require.False(t, shouldPrintUsageHint(nil))
}

func TestHelpHintTarget(t *testing.T) {
	t.Parallel()

	root := cli.NewRootCmd()
	require.Equal(t, "voxclip", helpHintTarget(root, []string{"--badflag"}))
	require.Equal(t, "voxclip", helpHintTarget(root, []string{"badcmd"}))
	require.Equal(t, "voxclip transcribe", helpHintTarget(root, []string{"transcribe"}))
	require.Equal(t, "voxclip transcribe", helpHintTarget(root, []string{"transcribe", "--copy"}))
}
