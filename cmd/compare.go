package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kuferek/process"
	"log"
)

var left bool
var right bool

var cmdCompare = &cobra.Command{
	Use:               "compare [directory1] [directory2]",
	Short:             "Compares two directories by checksums",
	DisableAutoGenTag: true,
	Args:              cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir1 := args[0]
		dir2 := args[1]
		log.Printf("# Comparing: %s - %s", dir1, dir2)
		comparison, err := process.Compare(dir1, dir2, verify)
		if err != nil {
			return err
		}

		if comparison != nil {
			if left {
				for _, path := range comparison.Dir1 {
					fmt.Printf("> %s\n", path)
				}
			}
			if right {
				for _, path := range comparison.Dir2 {
					fmt.Printf("< %s\n", path)
				}
			}
		}

		return err
	},
}

func init() {
	cmdRoot.AddCommand(cmdCompare)
	cmdCompare.PersistentFlags().BoolVar(&verify, "verify", false, "force verify checksums")
	cmdCompare.PersistentFlags().BoolVarP(&left, "left", "1", false, "Show only files unique to left dir")
	cmdCompare.PersistentFlags().BoolVarP(&right, "right", "3", false, "Show only files unique to right dir")
}
