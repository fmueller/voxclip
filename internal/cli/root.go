package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fmueller/voxclip/internal/audio"
	"github.com/fmueller/voxclip/internal/clipboard"
	"github.com/fmueller/voxclip/internal/logging"
	"github.com/fmueller/voxclip/internal/platform"
	"github.com/fmueller/voxclip/internal/version"
	"github.com/fmueller/voxclip/internal/whisper"
	"go.uber.org/zap"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

type appState struct {
	verbose      bool
	jsonLogs     bool
	noProgress   bool
	model        string
	modelDir     string
	language     string
	autoDownload bool
	backend      string
	input        string
	inputFormat  string
	copyEmpty    bool
	silenceGate  bool
	silenceDBFS  float64
	duration     time.Duration
	immediate    bool

	logger *zap.Logger
	now    func() time.Time
	out    io.Writer

	preflightFn  func(ctx context.Context) error
	recordFn     func(ctx context.Context, opts recordOptions) (string, error)
	transcribeFn func(ctx context.Context, audioPath string) (string, error)
	copyFn       func(ctx context.Context, value string) error
}

func NewRootCmd() *cobra.Command {
	app := &appState{
		model:        "small",
		language:     "auto",
		autoDownload: true,
		backend:      "auto",
		silenceGate:  true,
		silenceDBFS:  -65,
		now:          time.Now,
		out:          os.Stdout,
	}
	app.preflightFn = app.ensureTranscriptionReady
	app.recordFn = app.recordAudio
	app.transcribeFn = app.transcribeAudio
	app.copyFn = clipboard.CopyText

	cmd := &cobra.Command{
		Use:           "voxclip",
		Short:         "Record and transcribe audio with a bundled whisper engine",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			logger, err := logging.New(logging.Options{Verbose: app.verbose, JSON: app.jsonLogs})
			if err != nil {
				return fmt.Errorf("initialize logger: %w", err)
			}
			app.language = sanitizeLanguage(app.language)
			app.logger = logger
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.runDefault(cmd.Context())
		},
	}

	cmd.PersistentFlags().BoolVar(&app.verbose, "verbose", false, "Enable verbose logs")
	cmd.PersistentFlags().BoolVar(&app.jsonLogs, "json", false, "Enable JSON logging")
	cmd.PersistentFlags().BoolVar(&app.noProgress, "no-progress", false, "Disable progress indicators")
	cmd.PersistentFlags().StringVar(&app.model, "model", "small", "Model name or model file path")
	cmd.PersistentFlags().StringVar(&app.modelDir, "model-dir", "", "Directory where models are stored")
	cmd.PersistentFlags().StringVar(&app.language, "language", "auto", "Language code (auto|en|de|...) for transcription")
	cmd.PersistentFlags().BoolVar(&app.autoDownload, "auto-download", true, "Automatically download missing models")
	cmd.PersistentFlags().StringVar(&app.backend, "backend", "auto", "Recording backend: auto|pw-record|arecord|ffmpeg")
	cmd.PersistentFlags().StringVar(&app.input, "input", "", "Input device selector for backend-specific usage")
	cmd.PersistentFlags().StringVar(&app.inputFormat, "input-format", "", "Input format for ffmpeg backend (pulse|alsa)")
	cmd.PersistentFlags().BoolVar(&app.copyEmpty, "copy-empty", false, "Copy blank transcripts to clipboard")
	cmd.PersistentFlags().BoolVar(&app.silenceGate, "silence-gate", true, "Detect near-silent WAV audio and skip transcription")
	cmd.PersistentFlags().Float64Var(&app.silenceDBFS, "silence-threshold-dbfs", -65, "Silence gate threshold in dBFS")
	cmd.Flags().DurationVar(&app.duration, "duration", 0, "Record duration, e.g. 10s; 0 means interactive start/stop")
	cmd.Flags().BoolVar(&app.immediate, "immediate", false, "Start recording immediately without waiting for Enter")

	cmd.AddCommand(newRecordCmd(app))
	cmd.AddCommand(newTranscribeCmd(app))
	cmd.AddCommand(newDevicesCmd(app))
	cmd.AddCommand(newSetupCmd(app))

	return cmd
}

func (a *appState) ensureTranscriptionReady(ctx context.Context) error {
	if _, err := whisper.NewBundledEngine(a.log()); err != nil {
		return err
	}
	if _, err := a.ensureModelAvailable(ctx); err != nil {
		return err
	}
	return nil
}

func (a *appState) runDefault(ctx context.Context) error {
	preflightFn := a.preflightFn
	if preflightFn == nil {
		preflightFn = a.ensureTranscriptionReady
	}

	recordFn := a.recordFn
	if recordFn == nil {
		recordFn = a.recordAudio
	}

	transcribeFn := a.transcribeFn
	if transcribeFn == nil {
		transcribeFn = a.transcribeAudio
	}

	copyFn := a.copyFn
	if copyFn == nil {
		copyFn = clipboard.CopyText
	}

	if err := preflightFn(ctx); err != nil {
		return err
	}

	audioPath, err := recordFn(ctx, recordOptions{duration: a.duration, input: a.input, format: a.inputFormat})
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(audioPath); err != nil {
			a.log().Warn("failed to remove recording", zap.String("path", audioPath), zap.Error(err))
		}
	}()

	transcript, skipped, err := a.silenceGateTranscript(audioPath)
	if err != nil {
		return err
	}
	if !skipped {
		transcript, err = transcribeFn(ctx, audioPath)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(a.outWriter(), transcript)
	if isBlankTranscript(transcript) {
		a.log().Warn(noSpeechHint())
		if !a.copyEmpty {
			return nil
		}
	}

	if err := copyFn(ctx, transcript); err != nil {
		if errors.Is(err, clipboard.ErrUnavailable) {
			a.log().Warn("clipboard tool unavailable; transcript left on stdout")
			return nil
		}
		a.log().Warn("failed to copy transcript to clipboard; transcript left on stdout", zap.Error(err))
		return nil
	}

	a.log().Info("transcript copied to clipboard")
	return nil
}

func (a *appState) modelStorageDir() (string, error) {
	dir, err := platform.ResolveModelDir(a.modelDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create model directory %s: %w", dir, err)
	}
	return dir, nil
}

func (a *appState) recordingOutputPath(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		if err := os.MkdirAll(filepath.Dir(override), 0o755); err != nil {
			return "", fmt.Errorf("create output directory: %w", err)
		}
		return override, nil
	}

	recordingDir, err := platform.ResolveRecordingDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(recordingDir, 0o755); err != nil {
		return "", fmt.Errorf("create recording directory %s: %w", recordingDir, err)
	}

	return filepath.Join(recordingDir, fmt.Sprintf("recording-%s.wav", a.now().Format("20060102-150405"))), nil
}

func (a *appState) log() *zap.Logger {
	if a.logger == nil {
		return zap.NewNop()
	}
	return a.logger
}

func (a *appState) progressEnabled() bool {
	if a.noProgress {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func (a *appState) outWriter() io.Writer {
	if a.out == nil {
		return os.Stdout
	}
	return a.out
}

func (a *appState) silenceGateTranscript(audioPath string) (string, bool, error) {
	if !a.silenceGate {
		return "", false, nil
	}

	if !strings.EqualFold(filepath.Ext(audioPath), ".wav") {
		return "", false, nil
	}

	silent, metrics, err := audio.IsSilentWAV(audioPath, a.silenceDBFS)
	if err != nil {
		a.log().Warn("silence gate analysis failed; continuing transcription", zap.Error(err), zap.String("audio", audioPath))
		return "", false, nil
	}

	if !silent {
		return "", false, nil
	}

	a.log().Info(
		"audio considered silent; skipping transcription",
		zap.String("audio", audioPath),
		zap.Float64("rms_dbfs", metrics.RMSdBFS),
		zap.Float64("peak_dbfs", metrics.PeakdBFS),
		zap.Float64("threshold_dbfs", a.silenceDBFS),
	)

	return blankAudioToken, true, nil
}
