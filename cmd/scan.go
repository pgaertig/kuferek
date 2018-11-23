package cmd

import (
	"github.com/spf13/cobra"
	"kuferek/process"
	"log"
)

var verify bool

var cmdScan = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan directory and calculate checksums",
	Long: `
The "scan" command scans directory of files and calculates their checksums (SHA256).
`,
	DisableAutoGenTag: true,
	Args:              cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		log.Printf("# Scanning: %s", dir)
		counter, err := process.ScanDir(dir, verify)
		log.Printf("# Scanned: %s (%d files)", dir, counter)
		return err
	},
}

func init() {
	cmdRoot.AddCommand(cmdScan)
	cmdScan.PersistentFlags().BoolVar(&verify, "verify", false, "force verify checksums")
}
