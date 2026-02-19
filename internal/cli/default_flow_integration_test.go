//go:build integration

package cli

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunDefaultFlowSuccess(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)

	app := &appState{
		out: out,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return "/tmp/voxclip-audio.wav", nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "hello world", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "hello world\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:/tmp/voxclip-audio.wav",
		"copy:hello world",
	}, order)
}

func TestRunDefaultClipboardFailureIsNonFatal(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)

	app := &appState{
		out: out,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return "/tmp/voxclip-audio.wav", nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "clipboard fallback", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return errors.New("clipboard command failed")
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "clipboard fallback\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:/tmp/voxclip-audio.wav",
		"copy:clipboard fallback",
	}, order)
}

func TestRunDefaultSkipsCopyForBlankTranscript(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)

	app := &appState{
		out: out,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return "/tmp/voxclip-audio.wav", nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "[BLANK_AUDIO]", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:/tmp/voxclip-audio.wav",
	}, order)
}

func TestRunDefaultCopiesBlankWhenCopyEmptyEnabled(t *testing.T) {
	var order []string
	out := new(bytes.Buffer)

	app := &appState{
		out:       out,
		copyEmpty: true,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			order = append(order, "record")
			return "/tmp/voxclip-audio.wav", nil
		},
		transcribeFn: func(_ context.Context, audioPath string) (string, error) {
			order = append(order, "transcribe:"+audioPath)
			return "[BLANK_AUDIO]", nil
		},
		copyFn: func(_ context.Context, value string) error {
			order = append(order, "copy:"+value)
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
	require.Equal(t, []string{
		"record",
		"transcribe:/tmp/voxclip-audio.wav",
		"copy:[BLANK_AUDIO]",
	}, order)
}

func TestRunDefaultSkipsTranscribeWhenRecordingIsSilent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "silent.wav")
	require.NoError(t, os.WriteFile(path, makePCM16WAVForIntegration(make([]int16, 16000), 16000, 1), 0o644))

	out := new(bytes.Buffer)
	transcribeCalls := 0
	copyCalls := 0

	app := &appState{
		out:         out,
		silenceGate: true,
		silenceDBFS: -65,
		recordFn: func(_ context.Context, _ recordOptions) (string, error) {
			return path, nil
		},
		transcribeFn: func(_ context.Context, _ string) (string, error) {
			transcribeCalls++
			return "should-not-happen", nil
		},
		copyFn: func(_ context.Context, _ string) error {
			copyCalls++
			return nil
		},
	}

	err := app.runDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, transcribeCalls)
	require.Equal(t, 0, copyCalls)
	require.Equal(t, "[BLANK_AUDIO]\n", out.String())
}

func makePCM16WAVForIntegration(samples []int16, sampleRate int, channels int) []byte {
	bytesPerSample := 2
	dataSize := len(samples) * bytesPerSample
	fmtChunkSize := 16
	riffSize := 4 + (8 + fmtChunkSize) + (8 + dataSize)

	out := make([]byte, 12+8+fmtChunkSize+8+dataSize)
	off := 0

	copy(out[off:], []byte("RIFF"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(riffSize))
	off += 4
	copy(out[off:], []byte("WAVE"))
	off += 4

	copy(out[off:], []byte("fmt "))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(fmtChunkSize))
	off += 4
	binary.LittleEndian.PutUint16(out[off:], 1)
	off += 2
	binary.LittleEndian.PutUint16(out[off:], uint16(channels))
	off += 2
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate*channels*bytesPerSample))
	off += 4
	binary.LittleEndian.PutUint16(out[off:], uint16(channels*bytesPerSample))
	off += 2
	binary.LittleEndian.PutUint16(out[off:], 16)
	off += 2

	copy(out[off:], []byte("data"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(dataSize))
	off += 4

	for _, s := range samples {
		binary.LittleEndian.PutUint16(out[off:], uint16(s))
		off += 2
	}

	return out
}
