package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/wormatter/pkg/formatter"
)

func init() {
	rootCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check if files need formatting (exit 1 if changes needed)")
	rootCmd.Flags().StringArrayVarP(&excludePatterns, "exclude", "e", nil, "Exclude files matching glob pattern (can be specified multiple times)")
}

var (
	excludePatterns []string
	rootCmd         = &cobra.Command{
		Use:     "wormatter <path>...",
		Short:   "A highly opinionated Go source code formatter",
		Long:    "Wormatter is a DST-based Go source code formatter. Highly opinionated, but very comprehensive. Gofumpt built-in.",
		Version: version,
		Args:    cobra.MinimumNArgs(1),
		RunE:    run,
	}
	version = "dev"

	checkOnly bool
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(_ *cobra.Command, args []string) error {
	opts := formatter.Options{
		CheckOnly:       checkOnly,
		ExcludePatterns: excludePatterns,
	}

	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("cannot access %q: %w", path, err)
		}

		if info.IsDir() {
			if err := formatter.FormatDirectory(path, opts); err != nil {
				return err
			}
		} else {
			if err := formatter.FormatFile(path, opts); err != nil {
				return err
			}
		}
	}

	return nil
}
