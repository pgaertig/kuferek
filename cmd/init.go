package cmd

import (
	"github.com/spf13/cobra"
	"kuferek/process"
	"log"
)

var cmdInit = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize directory as repository",
	Long: `
The "init" command initializes a new repository directory.
`,
	DisableAutoGenTag: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		log.Printf("# Initializing repo: %s", dir)
		err := process.InitRepo(dir, false)
		return err
	},
}

func init() {
	cmdRoot.AddCommand(cmdInit)
}
