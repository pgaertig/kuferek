package process

import (
	"fmt"
	"os"
	"path/filepath"
)

func ScanDir(dir string, verify bool) (int, error) {
	counter := 0
	fmt.Printf("#Scanning: %s\n", dir)

	if err := EnsureRepo(dir); err != nil {
		return 0, err
	}

	visit := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() && path != dir {
			fmt.Printf("#At: %s (processed %d)\n", path, counter)
			//dcnt, _ := ScanDir(path, verify)
			//counter += dcnt
			//return filepath.SkipDir
		}
		if f.Mode().IsRegular() {
			scanFile(path, f, verify)
			counter += 1
		}
		return nil
	}

	return counter, filepath.Walk(dir, visit)
}

func scanFile(path string, f os.FileInfo, verify bool) (sha256 string) {
	meta := readMeta(path)

	if meta == "" {
		meta = updateMeta(path, f)
		fmt.Printf("+ %s %s\n", meta, path)
	} else {
		validMeta := validateMeta(meta, f, path, verify)
		if validMeta != meta {
			fmt.Printf("- %s %s\n", meta, path)
			fmt.Printf("+ %s %s\n", validMeta, path)
			meta = validMeta
		} else {
			fmt.Printf("= %s %s\n", meta, path)
		}
	}

	return getMetaSha256(meta)
}
