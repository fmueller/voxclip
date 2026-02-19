package record

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type alsaBackend struct{}

func newALSARecorderBackend() Backend {
	return &alsaBackend{}
}

func (b *alsaBackend) Name() string {
	return "arecord"
}

func (b *alsaBackend) Available() bool {
	return commandAvailable("arecord")
}

func (b *alsaBackend) Record(ctx context.Context, cfg Config) error {
	if cfg.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	if err := os.MkdirAll(filepathDir(cfg.OutputPath), 0o755); err != nil {
		return err
	}

	args := []string{"-f", "S16_LE", "-r", strconv.Itoa(defaultSampleRate(cfg.SampleRate)), "-c", strconv.Itoa(defaultChannels(cfg.Channels)), cfg.OutputPath}
	if cfg.Duration > 0 {
		args = append([]string{"-d", strconv.Itoa(int(cfg.Duration / time.Second))}, args...)
	}

	cmd := exec.CommandContext(ctx, "arecord", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if cfg.Interactive {
		return runInteractiveCommand(ctx, cmd, cfg.Logger)
	}

	return cmd.Run()
}

func (b *alsaBackend) ListDevices(ctx context.Context) (string, error) {
	return commandOutput(ctx, "arecord", "-L")
}
