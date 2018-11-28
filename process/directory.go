package process

import (
	"fmt"
	"github.com/pkg/xattr"
	"os"
)

const repoXattr = "user.kuferek-repo"
const modeMaster = "master"
const modeCopy = "copy"

type pathError struct {
	path    string
	problem string
}

func (e *pathError) Error() string {
	return fmt.Sprintf("%s: %s", e.problem, e.path)
}

func InitRepo(dir string, master bool) error {
	if err := EnsureDir(dir); err != nil {
		return err
	}

	mode := "copy"
	if master {
		mode = "master"
	}

	xattr.Set(dir, repoXattr, []byte(mode))
	return nil
}

func EnsureDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("EnsureDir: can't stat '%s': %s", dir, err)
	}
	if !stat.IsDir() {
		return &pathError{dir, "Path is not a directory"}
	}
	return nil
}

func EnsureRepo(dir string) error {
	if err := EnsureDir(dir); err != nil {
		return err
	}

	val, err := xattr.Get(dir, repoXattr)
	sval := string(val)

	if err != nil || sval != modeCopy && sval != modeMaster {
		return &pathError{dir, "Path is not repo: "}
	}

	return nil
}

func EnsureDifferentRepos(dir1 string, dir2 string) error {
	if err := EnsureRepo(dir1); err != nil {
		return err
	}
	if err := EnsureRepo(dir2); err != nil {
		return err
	}

	stat1, _ := os.Stat(dir1)
	stat2, _ := os.Stat(dir2)

	if os.SameFile(stat1, stat2) {
		return &pathError{dir2, "Repos are the same directory"}
	}
	return nil
}
