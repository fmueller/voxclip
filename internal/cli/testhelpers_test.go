package cli

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func runCommand(t *testing.T, args []string) (stdout string, stderr string, err error) {
	t.Helper()

	cmd := NewRootCmd()
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

func makePCM16WAVForTest(samples []int16, sampleRate int, channels int) []byte {
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
