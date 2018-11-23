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
		}
		if f.Mode().IsRegular() {
			scanFile(path, f, verify)
			counter += 1
		}
		return nil
	}

	return counter, filepath.Walk(dir, visit)
}

type HashItem struct {
	hash string
	path string
	size int64
}

type Comparison struct {
	Dir1 []string
	Dir2 []string
}

type ScannedFileFunc func(path string, f os.FileInfo, hash string) error

func mapDir(dir string, verifyContent bool) (fileMap map[string]HashItem, err error) {
	fileList, err := listDir(dir, verifyContent)

	if err != nil {
		return
	}

	for _, item := range fileList {
		fileMap[item.hash] = item
	}

	return
}

func scanOneDir(dir string, verifyContent bool, fileFunc ScannedFileFunc) (err error) {
	fmt.Printf("# Scanning: %s\n", dir)
	visit := func(path string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			sha256 := scanFile(path, f, verifyContent)
			fileFunc(path, f, sha256)
		}
		return err
	}
	err = filepath.Walk(dir, visit)
	return
}

func listDir(dir string, verifyContent bool) (fileList []HashItem, err error) {
	err = scanOneDir(dir, verifyContent, func(path string, f os.FileInfo, hash string) error {
		fileList = append(fileList, HashItem{hash, path, f.Size()})
		return nil
	})
	return
}

func Compare(dir1 string, dir2 string, verifyContent bool) (comparison *Comparison, err error) {
	if err = EnsureDifferentRepos(dir1, dir2); err != nil {
		return
	}

	map1, err := mapDir(dir1, verifyContent)
	if err != nil {
		return
	}

	map2, err := mapDir(dir2, verifyContent)
	if err != nil {
		return
	}

	var list1 []string
	var list2 []string

	for k, v := range map1 {
		if _, ok := map2[k]; !ok {
			list1 = append(list1, v.path)
		}
	}

	for k, v := range map2 {
		if _, ok := map1[k]; !ok {
			list2 = append(list2, v.path)
		}
	}

	return &Comparison{list1, list2}, nil
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
