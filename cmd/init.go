package cmd

import (
	"github.com/spf13/cobra"
)

var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize master copy",
	Long: `
The "init" command initializes a new master copy.
`,
	DisableAutoGenTag: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmdRoot.AddCommand(cmdInit)
}