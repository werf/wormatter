package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/wormatter/pkg/formatter"
)

var (
	rootCmd = &cobra.Command{
		Use:     "wormatter <path>...",
		Short:   "A highly opinionated Go source code formatter",
		Long:    "Wormatter is a DST-based Go source code formatter. Highly opinionated, but very comprehensive. Gofumpt built-in.",
		Version: version,
		Args:    cobra.MinimumNArgs(1),
		RunE:    run,
	}
	version = "dev"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(_ *cobra.Command, args []string) error {
	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("cannot access %q: %w", path, err)
		}

		if info.IsDir() {
			if err := formatter.FormatDirectory(path); err != nil {
				return err
			}
		} else {
			if err := formatter.FormatFile(path); err != nil {
				return err
			}
		}
	}

	return nil
}
