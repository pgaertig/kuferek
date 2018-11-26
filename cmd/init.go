package cmd

import (
	"github.com/spf13/cobra"
	"kuferek/process"
	"log"
)

var cmdInit = &cobra.Command{
	Use:   "init <directory> [directory]...",
	Short: "Initialize directory as repository",
	Long: `
The "init" command initializes a new repository directory.
`,
	DisableAutoGenTag: true,
	Args:              cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (error) {
		for _, dir := range args {
			log.Printf("# Initializing repo: %s", dir)
			if err := process.InitRepo(dir, false); err != nil {
				log.Fatal(err)
			}
		}
		return nil
	},
}

func init() {
	cmdRoot.AddCommand(cmdInit)
}
