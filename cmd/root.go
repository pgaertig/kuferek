package cmd

import (
	"github.com/spf13/cobra"
)

var master string
var debug bool

// cmdRoot is the base command when no other command has been specified.
var cmdRoot = &cobra.Command{
	Use:   "kuferek",
	Short: "File directories synchronization and deduplication ",
	Long:  `Kuferek synchronizes and deduplicates file directories`,
}

func init() {
	cmdRoot.PersistentFlags().StringVarP(&master, "master", "m", ".", "master copy (default is current directory)")
	cmdRoot.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
}

func Start() {
	cmdRoot.Execute()
}
