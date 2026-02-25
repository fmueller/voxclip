package cli

import (
	"fmt"

	"github.com/fmueller/voxclip/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "voxclip v%s\n", version.Resolve())
			return nil
		},
	}
}
