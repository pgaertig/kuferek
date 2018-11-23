package process

import "os"

type DiskUsageStats struct {
	Real   int64
	Dedup  int64
	Count  int
	Unique int
}

func DiskUsage(dir string, verifyContent bool) (stats DiskUsageStats, err error) {
	if err = EnsureRepo(dir); err != nil {
		return
	}

	countHistory := make(map[string]bool)

	scanOneDir(dir, verifyContent, func(path string, f os.FileInfo, hash string) error {
		_, counted := countHistory[hash]
		if !counted {
			stats.Dedup += f.Size()
			stats.Unique += 1
			countHistory[hash] = true
		}
		stats.Real += f.Size()
		stats.Count += 1
		return nil
	})
	return
}
