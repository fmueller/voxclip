package cli

import (
	"fmt"
	"path/filepath"

	"github.com/fmueller/voxclip/internal/download"
	"github.com/fmueller/voxclip/internal/whisper"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newSetupCmd(app *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Download and verify speech model assets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			modelDir, err := app.modelStorageDir()
			if err != nil {
				return err
			}

			resolved, err := whisper.ResolveModel(app.model, modelDir)
			if err != nil {
				return err
			}
			if resolved.IsCustomPath {
				return fmt.Errorf("setup expects a named model; got custom path %s", resolved.Path)
			}

			expectedChecksum := resolved.SHA256
			if expectedChecksum == "" && resolved.SHA256URL != "" {
				checksum, err := download.ResolveExpectedChecksum(cmd.Context(), resolved.SHA256URL, filepath.Base(resolved.Path), nil)
				if err != nil {
					return fmt.Errorf("resolve checksum for model %s: %w", resolved.Name, err)
				}
				expectedChecksum = checksum
			}

			if !resolved.NeedsDownload {
				if expectedChecksum != "" {
					if err := download.VerifyFileChecksum(resolved.Path, expectedChecksum); err != nil {
						app.log().Warn("model checksum verification failed; downloading fresh copy", zap.String("model", resolved.Name), zap.Error(err))
						resolved.NeedsDownload = true
					}
				}
			}

			if !resolved.NeedsDownload {
				app.log().Info("model already present", zap.String("model", resolved.Name), zap.String("path", resolved.Path))
				fmt.Fprintf(cmd.OutOrStdout(), "Model %s already present at %s\n", resolved.Name, resolved.Path)
				return nil
			}

			app.log().Info("downloading model", zap.String("model", resolved.Name), zap.String("path", resolved.Path))
			if err := download.DownloadFile(cmd.Context(), download.Options{
				URL:            resolved.URL,
				Destination:    resolved.Path,
				ExpectedSHA256: expectedChecksum,
				ChecksumURL:    resolved.SHA256URL,
				NoProgress:     app.noProgress,
				Logger:         app.log(),
			}); err != nil {
				return fmt.Errorf("download model %s: %w", resolved.Name, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Model %s installed at %s\n", resolved.Name, resolved.Path)
			return nil
		},
	}

	bindLoggingFlags(cmd, app)
	bindProgressFlag(cmd, app)
	bindModelFlags(cmd, app)

	return cmd
}
