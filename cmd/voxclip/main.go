package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fmueller/voxclip/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if shouldPrintUsageHint(err) {
			fmt.Fprintf(os.Stderr, "Run '%s --help' for usage.\n", helpHintTarget(cmd, os.Args[1:]))
		}
		os.Exit(1)
	}
}

func shouldPrintUsageHint(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	patterns := []string{
		"unknown command",
		"unknown flag",
		"unknown shorthand flag",
		"accepts ",
		"requires at least",
		"requires at most",
		"requires between",
		"required flag",
		"missing required",
	}

	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	return false
}

func helpHintTarget(root *cobra.Command, args []string) string {
	if root == nil {
		return "voxclip"
	}

	target := root.CommandPath()
	if len(args) == 0 {
		return target
	}

	if strings.HasPrefix(args[0], "-") {
		return target
	}

	found, _, err := root.Find(args)
	if err == nil && found != nil {
		return found.CommandPath()
	}

	return target
}
