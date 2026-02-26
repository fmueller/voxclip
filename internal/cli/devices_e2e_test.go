//go:build linux

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevicesEndToEndAllBackends(t *testing.T) {
	stubDir := t.TempDir()
	writeStub(t, stubDir, "pw-record", "#!/bin/sh\nexit 0\n")
	writeStub(t, stubDir, "pw-cli", "#!/bin/sh\necho 'id 42, type PipeWire:Interface:Node'\necho 'node.name = \"alsa_input.usb\"'\n")
	writeStub(t, stubDir, "arecord", stubScript("arecord", "default\nhw:0,0\n"))
	writeStub(t, stubDir, "ffmpeg", "#!/bin/sh\nexit 0\n")
	writeStub(t, stubDir, "pactl", "#!/bin/sh\necho '1\talsa_output.pci monitor'\n")

	t.Setenv("PATH", stubDir)

	stdout, _, err := runCommand(t, []string{"devices"})
	require.NoError(t, err)

	require.Contains(t, stdout, "== pw-record ==")
	require.Contains(t, stdout, "== arecord ==")
	require.Contains(t, stdout, "== ffmpeg ==")

	require.Contains(t, stdout, "alsa_input.usb")
	require.Contains(t, stdout, "hw:0,0")
	require.Contains(t, stdout, "alsa_output.pci monitor")
}

func TestDevicesEndToEndPartialAvailability(t *testing.T) {
	stubDir := t.TempDir()
	writeStub(t, stubDir, "arecord", stubScript("arecord", "default\nhw:0,0\n"))

	t.Setenv("PATH", stubDir)

	stdout, _, err := runCommand(t, []string{"devices"})
	require.NoError(t, err)

	require.Contains(t, stdout, "== pw-record ==")
	require.Contains(t, stdout, "not available on PATH")

	require.Contains(t, stdout, "== arecord ==")
	require.Contains(t, stdout, "hw:0,0")

	require.Contains(t, stdout, "== ffmpeg ==")
}

func TestDevicesEndToEndCommandFailure(t *testing.T) {
	stubDir := t.TempDir()
	writeStub(t, stubDir, "pw-record", "#!/bin/sh\nexit 0\n")
	writeStub(t, stubDir, "pw-cli", "#!/bin/sh\nexit 1\n")

	t.Setenv("PATH", stubDir)

	stdout, _, err := runCommand(t, []string{"devices"})
	require.NoError(t, err)

	require.Contains(t, stdout, "== pw-record ==")
	require.Contains(t, stdout, "failed to list devices")
}

func writeStub(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o755))
}

func stubScript(_, output string) string {
	return "#!/bin/sh\necho '" + output + "'\n"
}
