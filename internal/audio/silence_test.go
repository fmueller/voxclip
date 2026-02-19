package audio

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSilentWAVDetectsSilence(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "silent.wav")
	require.NoError(t, os.WriteFile(path, makePCM16WAV(make([]int16, 16000), 16000, 1), 0o644))

	silent, metrics, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
	require.True(t, silent)
	require.True(t, math.IsInf(metrics.RMSdBFS, -1))
	require.True(t, math.IsInf(metrics.PeakdBFS, -1))
	require.EqualValues(t, 16000, metrics.Samples)
}

func TestIsSilentWAVDetectsSpeechLikeSignal(t *testing.T) {
	t.Parallel()

	samples := make([]int16, 16000)
	for i := range samples {
		samples[i] = int16(0.25 * 32767 * math.Sin(2*math.Pi*440*float64(i)/16000.0))
	}

	path := filepath.Join(t.TempDir(), "voice.wav")
	require.NoError(t, os.WriteFile(path, makePCM16WAV(samples, 16000, 1), 0o644))

	silent, metrics, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
	require.False(t, silent)
	require.Greater(t, metrics.PeakdBFS, -20.0)
	require.Greater(t, metrics.RMSdBFS, -20.0)
}

func TestIsSilentWAVInvalidFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "not-wav.wav")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0o644))

	_, _, err := IsSilentWAV(path, -65)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWAV)
}

func makePCM16WAV(samples []int16, sampleRate int, channels int) []byte {
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
