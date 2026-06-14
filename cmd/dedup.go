package cmd

import (
	"github.com/spf13/cobra"
	"kuferek/process"
)

var dedupWrite = false
var dedupMedia = false

var cmdDedup = &cobra.Command{
	Use:   "dedup [--write|-w] [--media|-m] <master-dir>...",
	Short: "Move current-dir files already present in master out of the way",
	Long: `
The "dedup" command finds files under the current directory (recursively,
skipping "@eaDir" and its own "!found-in-master" directories) that already exist
somewhere under any of the given master dirs (compared by SHA256). Files smaller
than 1024 bytes and ignored junk files (e.g. Thumbs.db) are skipped, and a match
additionally requires the file extension to match (case-insensitive). With
--media only media files (images, video, audio) are considered. A file that
shares a name and size with a master file but has a different checksum is reported
as a possible bitrot ("checksum mismatch!") and counted, but not moved. With
--write it moves the content-identical files into the "!found-in-master"
subdirectory, preserving their relative path, so the space can be reclaimed.
Master dirs are read-only and never modified. Without --write the command only
reports what it would do (dry-run). One or more master dirs may be given, so shell
wildcards work, e.g. "kuferek dedup 202[1234]" or "kuferek dedup /backups/*".
`,
	DisableAutoGenTag: true,
	Args:              cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return process.Dedup(args, dedupWrite, dedupMedia)
	},
}

func init() {
	cmdRoot.AddCommand(cmdDedup)
	cmdDedup.Flags().BoolVarP(&dedupWrite, "write", "w", false, "move matched files (default is dry-run)")
	cmdDedup.Flags().BoolVarP(&dedupMedia, "media", "m", false, "only consider media files (images, video, audio)")
}
