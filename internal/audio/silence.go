package audio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

var (
	ErrUnsupportedWAV = errors.New("unsupported wav format")
	ErrInvalidWAV     = errors.New("invalid wav file")
)

type SilenceMetrics struct {
	RMSdBFS  float64
	PeakdBFS float64
	Samples  int64
}

func IsSilentWAV(path string, thresholdDBFS float64) (bool, SilenceMetrics, error) {
	metrics, err := analyzeWAV(path)
	if err != nil {
		return false, SilenceMetrics{}, err
	}

	if metrics.Samples == 0 {
		return true, metrics, nil
	}

	if math.IsInf(metrics.RMSdBFS, -1) && math.IsInf(metrics.PeakdBFS, -1) {
		return true, metrics, nil
	}

	peakGate := thresholdDBFS + 6
	return metrics.RMSdBFS <= thresholdDBFS && metrics.PeakdBFS <= peakGate, metrics, nil
}

func analyzeWAV(path string) (SilenceMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return SilenceMetrics{}, fmt.Errorf("open wav: %w", err)
	}
	defer f.Close()

	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return SilenceMetrics{}, fmt.Errorf("%w: %v", ErrInvalidWAV, err)
		}
		return SilenceMetrics{}, fmt.Errorf("read wav header: %w", err)
	}

	if string(header[:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return SilenceMetrics{}, ErrInvalidWAV
	}

	var (
		audioFormat   uint16
		bitsPerSample uint16
		dataOffset    int64
		dataSize      uint32
		hasFmt        bool
		hasData       bool
	)

	for {
		chunkHeader := make([]byte, 8)
		if _, err := io.ReadFull(f, chunkHeader); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return SilenceMetrics{}, fmt.Errorf("read wav chunk header: %w", err)
		}

		chunkID := string(chunkHeader[:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		chunkStart, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return SilenceMetrics{}, fmt.Errorf("seek wav chunk start: %w", err)
		}

		skip := int64(chunkSize)
		if chunkSize%2 != 0 {
			skip++
		}

		switch chunkID {
		case "fmt ":
			if chunkSize < 16 {
				return SilenceMetrics{}, ErrInvalidWAV
			}

			buf := make([]byte, chunkSize)
			if _, err := io.ReadFull(f, buf); err != nil {
				return SilenceMetrics{}, fmt.Errorf("read wav fmt chunk: %w", err)
			}

			audioFormat = binary.LittleEndian.Uint16(buf[0:2])
			bitsPerSample = binary.LittleEndian.Uint16(buf[14:16])
			hasFmt = true

			if chunkSize%2 != 0 {
				if _, err := f.Seek(1, io.SeekCurrent); err != nil {
					return SilenceMetrics{}, fmt.Errorf("seek wav fmt padding: %w", err)
				}
			}
		case "data":
			dataOffset = chunkStart
			dataSize = chunkSize
			hasData = true
			if _, err := f.Seek(skip, io.SeekCurrent); err != nil {
				return SilenceMetrics{}, fmt.Errorf("seek wav data chunk: %w", err)
			}
		default:
			if _, err := f.Seek(skip, io.SeekCurrent); err != nil {
				return SilenceMetrics{}, fmt.Errorf("seek wav chunk %s: %w", chunkID, err)
			}
		}
	}

	if !hasFmt || !hasData {
		return SilenceMetrics{}, ErrInvalidWAV
	}

	if err := validateFormat(audioFormat, bitsPerSample); err != nil {
		return SilenceMetrics{}, err
	}

	if _, err := f.Seek(dataOffset, io.SeekStart); err != nil {
		return SilenceMetrics{}, fmt.Errorf("seek wav data offset: %w", err)
	}

	data := make([]byte, dataSize)
	if _, err := io.ReadFull(f, data); err != nil {
		return SilenceMetrics{}, fmt.Errorf("read wav data: %w", err)
	}

	peak, sumSquares, samples, err := measureSamples(data, audioFormat, bitsPerSample)
	if err != nil {
		return SilenceMetrics{}, err
	}

	if samples == 0 {
		return SilenceMetrics{RMSdBFS: math.Inf(-1), PeakdBFS: math.Inf(-1), Samples: 0}, nil
	}

	rms := math.Sqrt(sumSquares / float64(samples))
	return SilenceMetrics{
		RMSdBFS:  amplitudeToDBFS(rms),
		PeakdBFS: amplitudeToDBFS(peak),
		Samples:  samples,
	}, nil
}

func validateFormat(audioFormat, bitsPerSample uint16) error {
	if audioFormat != 1 && audioFormat != 3 {
		return ErrUnsupportedWAV
	}

	if audioFormat == 1 {
		switch bitsPerSample {
		case 8, 16, 24, 32:
			return nil
		default:
			return ErrUnsupportedWAV
		}
	}

	if audioFormat == 3 {
		switch bitsPerSample {
		case 32, 64:
			return nil
		default:
			return ErrUnsupportedWAV
		}
	}

	return ErrUnsupportedWAV
}

func measureSamples(data []byte, audioFormat, bitsPerSample uint16) (float64, float64, int64, error) {
	bytesPerSample := int(bitsPerSample / 8)
	if bytesPerSample <= 0 {
		return 0, 0, 0, ErrUnsupportedWAV
	}

	var peak float64
	var sumSquares float64
	var samples int64

	for i := 0; i+bytesPerSample <= len(data); i += bytesPerSample {
		value, err := decodeSample(data[i:i+bytesPerSample], audioFormat, bitsPerSample)
		if err != nil {
			return 0, 0, 0, err
		}

		abs := math.Abs(value)
		if abs > peak {
			peak = abs
		}
		sumSquares += value * value
		samples++
	}

	return peak, sumSquares, samples, nil
}

func decodeSample(sample []byte, audioFormat, bitsPerSample uint16) (float64, error) {
	if audioFormat == 3 {
		switch bitsPerSample {
		case 32:
			bits := binary.LittleEndian.Uint32(sample)
			return float64(math.Float32frombits(bits)), nil
		case 64:
			bits := binary.LittleEndian.Uint64(sample)
			return math.Float64frombits(bits), nil
		default:
			return 0, ErrUnsupportedWAV
		}
	}

	switch bitsPerSample {
	case 8:
		u := float64(sample[0])
		return (u - 128.0) / 128.0, nil
	case 16:
		v := int16(binary.LittleEndian.Uint16(sample))
		return float64(v) / 32768.0, nil
	case 24:
		v := int32(sample[0]) | int32(sample[1])<<8 | int32(sample[2])<<16
		if v&0x800000 != 0 {
			v |= ^0xFFFFFF
		}
		return float64(v) / 8388608.0, nil
	case 32:
		v := int32(binary.LittleEndian.Uint32(sample))
		return float64(v) / 2147483648.0, nil
	default:
		return 0, ErrUnsupportedWAV
	}
}

func amplitudeToDBFS(amplitude float64) float64 {
	if amplitude <= 0 {
		return math.Inf(-1)
	}
	return 20.0 * math.Log10(amplitude)
}
