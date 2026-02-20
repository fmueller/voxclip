package record

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ffmpegLinuxBackend struct{}

func newFFMPEGLinuxBackend() Backend {
	return &ffmpegLinuxBackend{}
}

func (b *ffmpegLinuxBackend) Name() string {
	return "ffmpeg"
}

func (b *ffmpegLinuxBackend) Available() bool {
	return commandAvailable("ffmpeg")
}

func (b *ffmpegLinuxBackend) Record(ctx context.Context, cfg Config) error {
	if cfg.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	if err := os.MkdirAll(filepathDir(cfg.OutputPath), 0o755); err != nil {
		return err
	}

	formats := []struct {
		format string
		input  string
	}{
		{format: "pulse", input: "default"},
		{format: "alsa", input: "default"},
	}

	if cfg.Format != "" {
		input := cfg.Input
		if input == "" {
			input = "default"
		}
		formats = []struct {
			format string
			input  string
		}{{format: cfg.Format, input: input}}
	}

	var errs []error
	for _, candidate := range formats {
		args := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-y", "-f", candidate.format, "-i", candidate.input}
		if cfg.Duration > 0 {
			args = append(args, "-t", strconv.Itoa(int(cfg.Duration/time.Second)))
		}
		args = append(args,
			"-ac", strconv.Itoa(defaultChannels(cfg.Channels)),
			"-ar", strconv.Itoa(defaultSampleRate(cfg.SampleRate)),
			"-c:a", "pcm_s16le",
			cfg.OutputPath,
		)

		var cmd *exec.Cmd
		if cfg.Interactive {
			cmd = exec.CommandContext(ctx, "ffmpeg", args...)
		} else if cfg.Duration > 0 {
			cmd = exec.Command("ffmpeg", args...)
		} else {
			cmd = exec.CommandContext(ctx, "ffmpeg", args...)
		}
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		if cfg.Interactive {
			err := runInteractiveCommand(ctx, cmd, cfg.Logger)
			if err == nil {
				return nil
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			errs = append(errs, fmt.Errorf("ffmpeg (%s/%s): %w", candidate.format, candidate.input, err))
			continue
		}

		if cfg.Duration > 0 {
			err := runTimedCommand(ctx, cmd, cfg.Duration, cfg.Logger)
			if err == nil {
				return nil
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			errs = append(errs, fmt.Errorf("ffmpeg (%s/%s): %w", candidate.format, candidate.input, err))
			continue
		}

		if err := cmd.Run(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			errs = append(errs, fmt.Errorf("ffmpeg (%s/%s): %w", candidate.format, candidate.input, err))
			continue
		}

		return nil
	}

	return errors.Join(errs...)
}

func (b *ffmpegLinuxBackend) ListDevices(ctx context.Context) (string, error) {
	var sections []string

	if commandAvailable("pactl") {
		if out, err := commandOutput(ctx, "pactl", "list", "short", "sources"); err == nil {
			sections = append(sections, "PulseAudio/PipeWire sources:\n"+out)
		} else {
			sections = append(sections, "PulseAudio/PipeWire sources: "+err.Error())
		}
	}

	if commandAvailable("arecord") {
		if out, err := commandOutput(ctx, "arecord", "-L"); err == nil {
			sections = append(sections, "ALSA devices:\n"+out)
		} else {
			sections = append(sections, "ALSA devices: "+err.Error())
		}
	}

	if len(sections) == 0 {
		return "", errors.New("no device listing command available")
	}

	return strings.Join(sections, "\n\n"), nil
}

func defaultSampleRate(value int) int {
	if value <= 0 {
		return 16000
	}
	return value
}

func defaultChannels(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func filepathDir(path string) string {
	return filepath.Dir(filepath.Clean(path))
}
