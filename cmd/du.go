package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kuferek/process"
)

var cmdDu = &cobra.Command{
	Use:               "du <directory> [directory]...",
	Short:             "Show disk usage by directory, real and dedup",
	DisableAutoGenTag: true,
	Args:              cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := process.DiskUsage(args, false)
		fmt.Printf("Files count: %d\n", stats.Count)
		fmt.Printf("Unique files: %d\n", stats.Unique)
		fmt.Printf("Real usage: %d\n", stats.Real)
		fmt.Printf("Deduplicated usage: %d\n", stats.Dedup)

		return err
	},
}

func init() {
	cmdRoot.AddCommand(cmdDu)
}
