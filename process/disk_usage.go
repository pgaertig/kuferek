package process

import "os"

type DiskUsageStats struct {
	Real   int64
	Dedup  int64
	Count  int
	Unique int
}

func DiskUsage(dirs []string, verifyContent bool) (stats DiskUsageStats, err error) {

	for _, dir := range dirs {
		if err = EnsureRepo(dir); err != nil {
			return
		}
	}

	countHistory := make(map[string]bool)

	for _, dir := range dirs {
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
	}
	return
}
