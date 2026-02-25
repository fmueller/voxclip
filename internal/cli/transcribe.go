package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fmueller/voxclip/internal/clipboard"
	"github.com/fmueller/voxclip/internal/download"
	"github.com/fmueller/voxclip/internal/whisper"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newTranscribeCmd(app *appState) *cobra.Command {
	var copyToClipboard bool

	cmd := &cobra.Command{
		Use:   "transcribe <audio-file>",
		Short: "Transcribe an audio file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transcribeFn := app.transcribeFn
			if transcribeFn == nil {
				transcribeFn = app.transcribeAudio
			}

			copyFn := app.copyFn
			if copyFn == nil {
				copyFn = clipboard.CopyText
			}

			transcript, err := transcribeFn(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), transcript)
			if isBlankTranscript(transcript) {
				app.log().Warn(noSpeechHint())
			}
			if copyToClipboard {
				if isBlankTranscript(transcript) && !app.copyEmpty {
					return nil
				}

				if err := copyFn(cmd.Context(), transcript); err != nil {
					return err
				}
				app.log().Info("transcript copied to clipboard")
			}
			return nil
		},
	}

	bindLoggingFlags(cmd, app)
	bindProgressFlag(cmd, app)
	bindModelFlags(cmd, app)
	bindLanguageAndModelDownloadFlags(cmd, app)
	bindCopyAndSilenceFlags(cmd, app)
	cmd.Flags().BoolVar(&copyToClipboard, "copy", false, "Copy transcript to clipboard")
	return cmd
}

func (a *appState) transcribeAudio(ctx context.Context, audioPath string) (string, error) {
	audioPath = filepath.Clean(audioPath)
	if _, err := os.Stat(audioPath); err != nil {
		return "", fmt.Errorf("audio file not found: %w", err)
	}

	if transcript, skipped, err := a.silenceGateTranscript(audioPath); err != nil {
		return "", err
	} else if skipped {
		return transcript, nil
	}

	model, err := a.ensureModelAvailable(ctx)
	if err != nil {
		return "", err
	}

	engine, err := whisper.NewBundledEngine(a.log())
	if err != nil {
		return "", err
	}

	a.log().Info("transcribing...", zap.String("audio", audioPath), zap.String("model", model.Path), zap.String("language", a.language))
	stopSpinner := startSpinner(a.progressEnabled(), "Transcribing")
	started := time.Now()

	transcript, err := engine.Transcribe(ctx, whisper.TranscriptionRequest{
		AudioPath: audioPath,
		ModelPath: model.Path,
		Language:  a.language,
	})
	stopSpinner()
	if err != nil {
		a.log().Warn("transcription failed", zap.Duration("elapsed", time.Since(started)), zap.Error(err))
		return "", err
	}
	a.log().Info("transcription finished", zap.Duration("elapsed", time.Since(started)))

	return transcript, nil
}

func (a *appState) ensureModelAvailable(ctx context.Context) (whisper.ResolvedModel, error) {
	modelDir, err := a.modelStorageDir()
	if err != nil {
		return whisper.ResolvedModel{}, err
	}

	resolved, err := whisper.ResolveModel(a.model, modelDir)
	if err != nil {
		return whisper.ResolvedModel{}, err
	}

	if !resolved.NeedsDownload {
		return resolved, nil
	}

	if !a.autoDownload {
		return whisper.ResolvedModel{}, fmt.Errorf("model %q is missing at %s; run `voxclip setup --model %s` or use --auto-download=true", resolved.Name, resolved.Path, resolved.Name)
	}

	a.log().Info("model not found, downloading", zap.String("model", resolved.Name), zap.String("destination", resolved.Path))
	if err := download.DownloadFile(ctx, download.Options{
		URL:            resolved.URL,
		Destination:    resolved.Path,
		ExpectedSHA256: resolved.SHA256,
		ChecksumURL:    resolved.SHA256URL,
		NoProgress:     a.noProgress,
		Logger:         a.log(),
	}); err != nil {
		return whisper.ResolvedModel{}, fmt.Errorf("download model %q: %w", resolved.Name, err)
	}

	resolved.NeedsDownload = false
	return resolved, nil
}

func sanitizeLanguage(input string) string {
	trimmed := strings.TrimSpace(strings.ToLower(input))
	if trimmed == "" {
		return "auto"
	}
	return trimmed
}
