package cmd

import (
	"github.com/spf13/cobra"
	"kuferek/process"
	"log"
)


var verify bool

var cmdScan = &cobra.Command{
	Use:   "scan",
	Short: "Scan copy and calculate checksums",
	Long: `
The "scan" command scans directory for files and calculates their checksums (SHA256).
`,
	DisableAutoGenTag: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("#Scanning: %s", master)
		counter, err := process.ScanDir(master, verify)
		log.Printf("#Scanned: %s (%d files)", master, counter)
		return err
	},
}

func init() {
	cmdRoot.AddCommand(cmdScan)
	cmdScan.PersistentFlags().BoolVar(&verify, "verify", false, "force verify checksums")
}