package cli

import (
	"fmt"

	"go.uber.org/zap"
)

func writePIDFile(path string) error {
	return fmt.Errorf("--pid-file is not supported on Windows")
}

func removePIDFile(path string, logger *zap.Logger) {
	logger.Warn("--pid-file is not supported on Windows")
}

func registerStopSignal(stopCh chan struct{}, logger *zap.Logger) func() {
	logger.Warn("signal-based stop is not supported on Windows")
	return func() {}
}
