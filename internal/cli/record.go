package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fmueller/voxclip/internal/record"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type recordOptions struct {
	duration time.Duration
	output   string
	input    string
	format   string
}

func newRecordCmd(app *appState) *cobra.Command {
	opts := &recordOptions{}

	cmd := &cobra.Command{
		Use:   "record",
		Short: "Record audio into a WAV file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.input = app.input
			opts.format = app.inputFormat
			path, err := app.recordAudio(cmd.Context(), *opts)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}

	bindLoggingFlags(cmd, app)
	bindProgressFlag(cmd, app)
	bindRecordingBackendFlags(cmd, app)
	cmd.Flags().DurationVar(&opts.duration, "duration", 0, "Record duration, e.g. 6s; 0 means interactive start/stop (acts as max timeout with --pid-file)")
	cmd.Flags().StringVar(&opts.output, "output", "", "Output WAV file path")
	cmd.Flags().BoolVar(&app.immediate, "immediate", false, "Start recording immediately without waiting for Enter")
	cmd.Flags().StringVar(&app.pidFile, "pid-file", "", "Write PID to file and wait for SIGUSR1 to stop recording")

	return cmd
}

func (a *appState) recordAudio(ctx context.Context, opts recordOptions) (string, error) {
	outPath, err := a.recordingOutputPath(opts.output)
	if err != nil {
		return "", err
	}

	interactive := opts.duration <= 0
	if a.pidFile != "" {
		interactive = false
	}

	if interactive && !a.immediate {
		if err := record.WaitForEnter(os.Stdin, os.Stderr, "Press Enter to start recording."); err != nil {
			return "", err
		}
	}

	a.log().Info("recording started", zap.String("backend", a.backend), zap.String("output", outPath))
	stopProgress := func() {}
	if a.pidFile != "" {
		stopProgress = startSpinner(os.Stderr, a.progressEnabled(), "Recording")
	} else if interactive {
		stopProgress = startSpinner(os.Stderr, a.progressEnabled(), "Recording")
	} else {
		stopProgress = startDurationProgress(os.Stderr, a.progressEnabled(), "Recording", opts.duration)
	}
	defer stopProgress()

	recConfig := record.Config{
		OutputPath:  outPath,
		Duration:    opts.duration,
		Interactive: interactive,
		SampleRate:  16000,
		Channels:    1,
		Input:       opts.input,
		Format:      opts.format,
		Logger:      a.log(),
	}

	if a.pidFile != "" {
		stopCh := make(chan struct{})
		cleanup := registerStopSignal(stopCh, a.log())
		defer cleanup()

		if err := writePIDFile(a.pidFile); err != nil {
			return "", fmt.Errorf("write pid file: %w", err)
		}
		defer removePIDFile(a.pidFile, a.log())

		recConfig.StopCh = stopCh
	}

	backendName, err := record.RecordWithFallback(ctx, a.backend, recConfig)
	stopProgress()
	if err != nil {
		return "", err
	}

	a.log().Info("recording finished", zap.String("backend", backendName), zap.String("path", outPath))
	return outPath, nil
}
