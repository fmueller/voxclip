//go:build !windows

package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.uber.org/zap"
)

func writePIDFile(path string) error {
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

func removePIDFile(path string, logger *zap.Logger) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		logger.Warn("failed to remove pid file", zap.String("path", path), zap.Error(err))
	}
}

func registerStopSignal(stopCh chan struct{}, logger *zap.Logger) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)

	go func() {
		sig, ok := <-sigCh
		if !ok {
			return
		}
		logger.Debug(fmt.Sprintf("received %s; stopping recording", sig))
		close(stopCh)
	}()

	return func() {
		signal.Stop(sigCh)
		close(sigCh)
	}
}
