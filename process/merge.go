package process

import (
	"path/filepath"
	"os"
	"io"
)

type MergedItemFunc func (path string, merged bool, error error) error

func Merge(master string, dir string, target string, overwrite bool, verifyContent bool, mergedItemFunc MergedItemFunc) (err error) {

	if err = EnsureDifferentRepos(master, dir); err != nil {
		return
	}

	if err = EnsureDir(target); err != nil {
		return
	}

	map1, err := mapDir(master, verifyContent)
	if err != nil {
		return
	}

	list2, err := listDir(dir, verifyContent)
	if err != nil {
		return
	}

	for _, v := range list2 {
		if _, ok := map1[v.hash]; !ok {
			copied, itemErr := copyRelativePath(v.path, dir, target, overwrite)
			err = mergedItemFunc(v.path, copied, itemErr)
			if itemErr != nil {
				return itemErr
			}
		}
	}

	return
}

func copyRelativePath(sourcePath string, sourceDir string, targetDir string, overwrite bool) (copied bool, err error) {
	relPath, err := filepath.Rel(sourceDir, sourcePath)
	if err != nil {
		return
	}
	targetPath := filepath.Join(targetDir, relPath)

	//fmt.Printf("> Copying [%s] -> [%s]\n", sourcePath, targetPath)

	targetPathDir := filepath.Dir(targetPath)

	stat, err := os.Stat(targetPathDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(targetPathDir, os.ModePerm)
			if err != nil {
				return
			}
		} else {
			return
		}
	} else if !stat.IsDir() {
		return false, &pathError{targetPath, "Path is not a directory"}
	}

	_, err = os.Stat(targetPath)

	if os.IsNotExist(err) || overwrite {
		err = copyFile(sourcePath, targetPath)
		if err == nil {
			err = copyFileTimes(sourcePath, targetPath)
		}
		return err == nil, err
	}

	return
}


func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func copyFileTimes(src, dst string) (err error) {
	stat, err := os.Stat(src)

	if err == nil {
		err = os.Chtimes(dst, stat.ModTime(), stat.ModTime())
	}

	return
}