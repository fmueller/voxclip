package cli

import (
	"fmt"
	"runtime"

	"github.com/fmueller/voxclip/internal/record"
	"github.com/spf13/cobra"
)

func newDevicesCmd(app *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "devices",
		Short: "List recording devices and backend diagnostics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			backends := record.DefaultBackends(runtime.GOOS)
			if len(backends) == 0 {
				return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
			}

			for _, backend := range backends {
				fmt.Fprintf(cmd.OutOrStdout(), "== %s ==\n", backend.Name())
				if !backend.Available() {
					fmt.Fprintln(cmd.OutOrStdout(), "not available on PATH")
					fmt.Fprintln(cmd.OutOrStdout())
					continue
				}

				out, err := backend.ListDevices(cmd.Context())
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "failed to list devices: %v\n\n", err)
					continue
				}

				if out == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "no output")
					fmt.Fprintln(cmd.OutOrStdout())
					continue
				}

				fmt.Fprintln(cmd.OutOrStdout(), out)
				fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}
}
