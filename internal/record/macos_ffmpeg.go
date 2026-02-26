package record

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type ffmpegMacBackend struct{}

func newFFMPEGMacOSBackend() Backend {
	return &ffmpegMacBackend{}
}

func (b *ffmpegMacBackend) Name() string {
	return "ffmpeg"
}

func (b *ffmpegMacBackend) Available() bool {
	return commandAvailable("ffmpeg")
}

func (b *ffmpegMacBackend) Record(ctx context.Context, cfg Config) error {
	if cfg.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	if err := os.MkdirAll(filepathDir(cfg.OutputPath), 0o755); err != nil {
		return err
	}

	input := cfg.Input
	if input == "" {
		input = ":0"
	}

	args := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-y", "-f", "avfoundation", "-i", input}
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
		return runInteractiveCommand(ctx, cmd, cfg.Logger)
	}

	if cfg.Duration > 0 {
		return runTimedCommand(ctx, cmd, cfg.Duration, cfg.Logger)
	}

	return cmd.Run()
}

func (b *ffmpegMacBackend) ListDevices(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	out, _ := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return "", fmt.Errorf("ffmpeg returned no device output")
	}
	return trimmed, nil
}
