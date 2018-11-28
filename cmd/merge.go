package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kuferek/process"
	"log"
)

var mergeTargetDir = ""
var mergeOverwrite = false
var mergeForce = false

var cmdMerge = &cobra.Command{
	Use: "merge -m [masterdir] -t [destdir] [dir]",
	Short: "Copy dir and masterdir file structure difference into destdir.",
	Long: "Copies unique files from dir into destdir, but only these which don't exist in masterdir. Relative directory structure is retained.",
	DisableAutoGenTag: true,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir1 := args[0]
		log.Printf("# Merging %s into %s after comparing to %s", dir1, mergeTargetDir, master)

		mergedItemFunc := func(path string, copied bool, itemError error) (err error) {
			if copied {
				fmt.Printf("Merged: %s\n", path)
			} else {
				if itemError == nil {
					fmt.Printf("Skipped existing: %s\n", path)
				} else {
					fmt.Printf("Skipped by error: %s: %s\n", path, itemError)
					if !mergeForce {
						return itemError
					}
				}
			}
			return nil
		}

		err := process.Merge(master, dir1, mergeTargetDir, mergeOverwrite, verify, mergedItemFunc)

		return err
	},
}

func init() {
	cmdRoot.AddCommand(cmdMerge)
	cmdMerge.PersistentFlags().StringVarP(&master, "master", "m", ".", "master copy (default is current directory)")
	cmdMerge.PersistentFlags().StringVarP(&mergeTargetDir, "target", "t", "", "where to copy unique files")
	cmdMerge.PersistentFlags().BoolVarP(&mergeOverwrite, "overwrite", "o", false, "overwrite target files")
	cmdMerge.PersistentFlags().BoolVarP(&mergeForce, "force", "f", false, "skip errors and continue merge")
}
