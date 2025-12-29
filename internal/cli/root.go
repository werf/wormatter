package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/wormatter/pkg/formatter"
)

var version = "dev"

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "wormatter <file.go|directory>",
	Short:   "A highly opinionated Go source code formatter",
	Long:    "Wormatter is a DST-based Go source code formatter. Highly opinionated, but very comprehensive. Gofumpt built-in.",
	Version: version,
	Args:    cobra.ExactArgs(1),
	RunE:    run,
}

func run(_ *cobra.Command, args []string) error {
	path := args[0]
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access %q: %w", path, err)
	}

	if info.IsDir() {
		return formatter.FormatDirectory(path)
	}

	return formatter.FormatFile(path)
}
