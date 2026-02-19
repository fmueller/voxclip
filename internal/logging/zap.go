package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Verbose bool
	JSON    bool
}

func New(opts Options) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if opts.Verbose {
		level = zapcore.DebugLevel
	}

	cfg := zap.NewProductionConfig()
	if !opts.JSON {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = ""
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeCaller = nil
	}

	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.DisableStacktrace = !opts.Verbose

	if opts.JSON {
		cfg.Encoding = "json"
	} else {
		cfg.Encoding = "console"
	}

	return cfg.Build()
}
