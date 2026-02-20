package record

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type pipewireBackend struct{}

func newPipeWireBackend() Backend {
	return &pipewireBackend{}
}

func (b *pipewireBackend) Name() string {
	return "pw-record"
}

func (b *pipewireBackend) Available() bool {
	return commandAvailable("pw-record")
}

func (b *pipewireBackend) Record(ctx context.Context, cfg Config) error {
	if cfg.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	if err := os.MkdirAll(filepathDir(cfg.OutputPath), 0o755); err != nil {
		return err
	}

	args := []string{"--rate", strconv.Itoa(defaultSampleRate(cfg.SampleRate)), "--channels", strconv.Itoa(defaultChannels(cfg.Channels)), "--format", "s16", cfg.OutputPath}

	cmd := exec.CommandContext(ctx, "pw-record", args...)
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

func (b *pipewireBackend) ListDevices(ctx context.Context) (string, error) {
	if commandAvailable("pw-cli") {
		return commandOutput(ctx, "pw-cli", "ls", "Node")
	}

	if out, err := commandOutput(ctx, "pw-record", "--list-targets"); err == nil {
		return out, nil
	}

	if commandAvailable("pactl") {
		return commandOutput(ctx, "pactl", "list", "short", "sources")
	}

	return "", errors.New("no pipewire device listing command available")
}
