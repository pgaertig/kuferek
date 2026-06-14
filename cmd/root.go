package cmd

import (
	"github.com/spf13/cobra"
)

var master string
var debug bool
var excludes []string

// cmdRoot is the base command when no other command has been specified.
var cmdRoot = &cobra.Command{
	Use:   "kuferek",
	Short: "File directories synchronization and deduplication ",
	Long:  `Kuferek synchronizes and deduplicates file directories`,
}

func init() {
	cmdRoot.Flags()
	cmdRoot.PersistentFlags().StringVar(&master, "master", ".", "master copy (default is current directory)")
	cmdRoot.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
	cmdRoot.PersistentFlags().StringArrayVarP(&excludes, "excludes", "e", nil ,"debug output")
}

func Start() {
	cmdRoot.Execute()
}
