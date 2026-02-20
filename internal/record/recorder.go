package record

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/term"
)

var ErrInteractiveRequiresTTY = errors.New("interactive recording requires terminal input")
var ErrNoBackendAvailable = errors.New("no recording backend available")

type Config struct {
	OutputPath  string
	Duration    time.Duration
	Interactive bool
	SampleRate  int
	Channels    int
	Input       string
	Format      string
	Logger      *zap.Logger
}

type Backend interface {
	Name() string
	Available() bool
	Record(ctx context.Context, cfg Config) error
	ListDevices(ctx context.Context) (string, error)
}

func SelectBackend(backends []Backend, preferred string) (Backend, error) {
	if len(backends) == 0 {
		return nil, errors.New("no backends configured")
	}

	if preferred != "" && preferred != "auto" {
		for _, backend := range backends {
			if backend.Name() == preferred {
				if !backend.Available() {
					return nil, fmt.Errorf("requested backend %q is not available", preferred)
				}
				return backend, nil
			}
		}
		return nil, fmt.Errorf("unknown backend %q", preferred)
	}

	for _, backend := range backends {
		if backend.Available() {
			return backend, nil
		}
	}

	return nil, ErrNoBackendAvailable
}

func DefaultBackends(goos string) []Backend {
	switch goos {
	case "linux":
		return []Backend{newPipeWireBackend(), newALSARecorderBackend(), newFFMPEGLinuxBackend()}
	case "darwin":
		return []Backend{newFFMPEGMacOSBackend()}
	default:
		return nil
	}
}

func NewBackend(preferred string) (Backend, error) {
	backends := DefaultBackends(runtime.GOOS)
	if len(backends) == 0 {
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return SelectBackend(backends, preferred)
}

func RecordWithFallback(ctx context.Context, preferred string, cfg Config) (string, error) {
	backends := DefaultBackends(runtime.GOOS)
	if len(backends) == 0 {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return recordWithFallback(ctx, backends, preferred, cfg)
}

func recordWithFallback(ctx context.Context, backends []Backend, preferred string, cfg Config) (string, error) {
	orderedBackends, err := orderBackends(backends, preferred)
	if err != nil {
		return "", err
	}

	var errs []error
	for _, backend := range orderedBackends {
		if !backend.Available() {
			errs = append(errs, fmt.Errorf("%s: backend is not available", backend.Name()))
			continue
		}

		err := backend.Record(ctx, cfg)
		if err == nil {
			return backend.Name(), nil
		}

		if cleanupErr := removePartialRecording(cfg.OutputPath); cleanupErr != nil {
			errs = append(errs, fmt.Errorf("%s: cleanup partial recording %q: %w", backend.Name(), cfg.OutputPath, cleanupErr))
		}

		err = fmt.Errorf("%s: %w", backend.Name(), err)
		errs = append(errs, err)

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
	}

	if len(errs) == 0 {
		return "", ErrNoBackendAvailable
	}

	return "", fmt.Errorf("record audio with available backends: %w", errors.Join(errs...))
}

func orderBackends(backends []Backend, preferred string) ([]Backend, error) {
	if len(backends) == 0 {
		return nil, errors.New("no backends configured")
	}

	if preferred == "" || preferred == "auto" {
		return backends, nil
	}

	preferredIndex := -1
	for i, backend := range backends {
		if backend.Name() == preferred {
			preferredIndex = i
			break
		}
	}
	if preferredIndex == -1 {
		return nil, fmt.Errorf("unknown backend %q", preferred)
	}

	ordered := make([]Backend, 0, len(backends))
	ordered = append(ordered, backends[preferredIndex])
	for i, backend := range backends {
		if i == preferredIndex {
			continue
		}
		ordered = append(ordered, backend)
	}

	return ordered, nil
}

func removePartialRecording(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func WaitForEnter(in io.Reader, out io.Writer, message string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return ErrInteractiveRequiresTTY
	}

	if message != "" {
		if _, err := fmt.Fprintln(out, message); err != nil {
			return err
		}
	}

	reader := bufio.NewReader(in)
	_, err := reader.ReadString('\n')
	return err
}

func runInteractiveCommand(ctx context.Context, cmd *exec.Cmd, logger *zap.Logger) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := WaitForEnter(os.Stdin, os.Stderr, "Recording... press Enter to stop."); err != nil {
		_ = cmd.Process.Signal(os.Interrupt)
		_ = cmd.Wait()
		return err
	}

	stopSignalSent := cmd.Process.Signal(os.Interrupt) == nil
	err := cmd.Wait()
	if err == nil {
		return nil
	}

	if stopSignalSent {
		logger.Debug("recording process exited after stop signal", zap.Error(err))
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				logger.Debug("recording process stopped by signal", zap.String("signal", status.Signal().String()))
				return nil
			}
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return err
	}
}

func runTimedCommand(ctx context.Context, cmd *exec.Cmd, duration time.Duration, logger *zap.Logger) error {
	if duration <= 0 {
		return cmd.Run()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		select {
		case err := <-done:
			return err
		case <-timer.C:
			stopSignalSent := cmd.Process.Signal(os.Interrupt) == nil
			err := <-done
			if err == nil {
				return nil
			}

			if stopSignalSent {
				logger.Debug("recording process exited after timed stop signal", zap.Error(err))
				return nil
			}

			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.Signaled() {
						logger.Debug("recording process stopped by signal", zap.String("signal", status.Signal().String()))
						return nil
					}
				}
			}

			return err
		case <-ctx.Done():
			_ = cmd.Process.Signal(os.Interrupt)
			<-done
			return ctx.Err()
		}
	}
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func commandOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err != nil {
		if trimmed != "" {
			return "", fmt.Errorf("%s %s failed: %w (%s)", name, strings.Join(args, " "), err, trimmed)
		}
		return "", fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return trimmed, nil
}
