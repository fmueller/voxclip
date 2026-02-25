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

func TestIsSilentWAV_8BitPCM(t *testing.T) {
	t.Parallel()

	// 8-bit PCM: unsigned, 128 = silence, 255 = max positive
	samples := []byte{255, 255, 255, 255}
	wav := makeWAV(1, 8, 16000, 1, samples)

	path := filepath.Join(t.TempDir(), "8bit.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	_, metrics, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
	require.EqualValues(t, 4, metrics.Samples)
	require.Greater(t, metrics.PeakdBFS, -1.0)
}

func TestIsSilentWAV_24BitPCM(t *testing.T) {
	t.Parallel()

	// 24-bit PCM: little-endian signed, 3 bytes per sample
	// Sample 1: max positive 0x7FFFFF
	// Sample 2: negative -1000 = 0xFFFC18 (exercises sign extension)
	samples := make([]byte, 6)
	samples[0] = 0xFF
	samples[1] = 0xFF
	samples[2] = 0x7F
	samples[3] = 0x18
	samples[4] = 0xFC
	samples[5] = 0xFF

	wav := makeWAV(1, 24, 16000, 1, samples)
	path := filepath.Join(t.TempDir(), "24bit.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	_, metrics, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
	require.EqualValues(t, 2, metrics.Samples)
	require.Greater(t, metrics.PeakdBFS, -1.0)
}

func TestIsSilentWAV_32BitPCM(t *testing.T) {
	t.Parallel()

	// 32-bit PCM: little-endian signed int32, near max value
	samples := make([]byte, 8)
	binary.LittleEndian.PutUint32(samples[0:4], uint32(int32(2000000000)))
	binary.LittleEndian.PutUint32(samples[4:8], uint32(int32(2000000000)))

	wav := makeWAV(1, 32, 16000, 1, samples)
	path := filepath.Join(t.TempDir(), "32bit.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	_, metrics, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
	require.EqualValues(t, 2, metrics.Samples)
	require.Greater(t, metrics.PeakdBFS, -2.0)
}

func TestIsSilentWAV_Float32(t *testing.T) {
	t.Parallel()

	// IEEE float32: audioFormat=3, bitsPerSample=32
	samples := make([]byte, 8)
	binary.LittleEndian.PutUint32(samples[0:4], math.Float32bits(0.0))
	binary.LittleEndian.PutUint32(samples[4:8], math.Float32bits(0.5))

	wav := makeWAV(3, 32, 16000, 1, samples)
	path := filepath.Join(t.TempDir(), "float32.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	silent, metrics, err := IsSilentWAV(path, -10)
	require.NoError(t, err)
	require.False(t, silent)
	require.EqualValues(t, 2, metrics.Samples)
}

func TestIsSilentWAV_EmptyData(t *testing.T) {
	t.Parallel()

	wav := makeWAV(1, 16, 16000, 1, nil)
	path := filepath.Join(t.TempDir(), "empty.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	silent, metrics, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
	require.True(t, silent)
	require.EqualValues(t, 0, metrics.Samples)
}

func TestIsSilentWAV_OddFmtChunk(t *testing.T) {
	t.Parallel()

	// Build a WAV with an odd-sized fmt chunk (17 bytes) to exercise the padding path
	wav := makeWAVWithOddFmt(1, 16, 16000, 1, make([]byte, 4))
	path := filepath.Join(t.TempDir(), "oddfmt.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	_, _, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
}

func TestIsSilentWAV_UnknownChunkBeforeData(t *testing.T) {
	t.Parallel()

	// Build a WAV with a LIST chunk inserted between fmt and data
	wav := makeWAVWithExtraChunk(1, 16, 16000, 1, make([]byte, 4))
	path := filepath.Join(t.TempDir(), "listchunk.wav")
	require.NoError(t, os.WriteFile(path, wav, 0o644))

	_, _, err := IsSilentWAV(path, -65)
	require.NoError(t, err)
}

// makeWAV builds a minimal WAV file with the given audio format, bit depth, and raw sample bytes.
func makeWAV(audioFormat, bitsPerSample uint16, sampleRate, channels int, sampleData []byte) []byte {
	bytesPerSample := int(bitsPerSample / 8)
	fmtChunkSize := 16
	dataSize := len(sampleData)
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
	binary.LittleEndian.PutUint16(out[off:], audioFormat)
	off += 2
	binary.LittleEndian.PutUint16(out[off:], uint16(channels))
	off += 2
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate*channels*bytesPerSample))
	off += 4
	binary.LittleEndian.PutUint16(out[off:], uint16(channels*bytesPerSample))
	off += 2
	binary.LittleEndian.PutUint16(out[off:], bitsPerSample)
	off += 2

	copy(out[off:], []byte("data"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(dataSize))
	off += 4

	copy(out[off:], sampleData)

	return out
}

// makeWAVWithOddFmt builds a WAV with a 17-byte fmt chunk (odd size) to exercise padding.
func makeWAVWithOddFmt(audioFormat, bitsPerSample uint16, sampleRate, channels int, sampleData []byte) []byte {
	bytesPerSample := int(bitsPerSample / 8)
	fmtChunkSize := 17 // odd size
	dataSize := len(sampleData)
	// RIFF size: "WAVE" + (8+fmtChunkSize+1pad) + (8+dataSize)
	riffSize := 4 + (8 + fmtChunkSize + 1) + (8 + dataSize)

	out := make([]byte, 12+8+fmtChunkSize+1+8+dataSize)
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
	binary.LittleEndian.PutUint16(out[off:], audioFormat)
	off += 2
	binary.LittleEndian.PutUint16(out[off:], uint16(channels))
	off += 2
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate*channels*bytesPerSample))
	off += 4
	binary.LittleEndian.PutUint16(out[off:], uint16(channels*bytesPerSample))
	off += 2
	binary.LittleEndian.PutUint16(out[off:], bitsPerSample)
	off += 2
	out[off] = 0 // extra byte to make 17
	off++
	out[off] = 0 // padding byte
	off++

	copy(out[off:], []byte("data"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(dataSize))
	off += 4
	copy(out[off:], sampleData)

	return out
}

// makeWAVWithExtraChunk builds a WAV with a LIST chunk between fmt and data.
func makeWAVWithExtraChunk(audioFormat, bitsPerSample uint16, sampleRate, channels int, sampleData []byte) []byte {
	bytesPerSample := int(bitsPerSample / 8)
	fmtChunkSize := 16
	listData := []byte("test")
	listChunkSize := len(listData)
	dataSize := len(sampleData)
	riffSize := 4 + (8 + fmtChunkSize) + (8 + listChunkSize) + (8 + dataSize)

	out := make([]byte, 12+8+fmtChunkSize+8+listChunkSize+8+dataSize)
	off := 0

	copy(out[off:], []byte("RIFF"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(riffSize))
	off += 4
	copy(out[off:], []byte("WAVE"))
	off += 4

	// fmt chunk
	copy(out[off:], []byte("fmt "))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(fmtChunkSize))
	off += 4
	binary.LittleEndian.PutUint16(out[off:], audioFormat)
	off += 2
	binary.LittleEndian.PutUint16(out[off:], uint16(channels))
	off += 2
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(sampleRate*channels*bytesPerSample))
	off += 4
	binary.LittleEndian.PutUint16(out[off:], uint16(channels*bytesPerSample))
	off += 2
	binary.LittleEndian.PutUint16(out[off:], bitsPerSample)
	off += 2

	// LIST chunk (unknown chunk)
	copy(out[off:], []byte("LIST"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(listChunkSize))
	off += 4
	copy(out[off:], listData)
	off += listChunkSize

	// data chunk
	copy(out[off:], []byte("data"))
	off += 4
	binary.LittleEndian.PutUint32(out[off:], uint32(dataSize))
	off += 4
	copy(out[off:], sampleData)

	return out
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
