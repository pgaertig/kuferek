package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// minFileSize is the smallest source file dedup considers; smaller files are
// skipped (the space saved rarely justifies tracking them).
const minFileSize = 1024

// ignoredNames are file names dedup never processes (junk/metadata files).
// Keys are lower-case; matching is case-insensitive.
var ignoredNames = map[string]bool{
	"thumbs.db":   true, // Windows thumbnail cache
	"ehthumbs.db": true, // Windows (enhanced) thumbnail cache
	"desktop.ini": true, // Windows folder settings
	".ds_store":   true, // macOS Finder metadata
}

func isIgnored(name string) bool {
	return ignoredNames[strings.ToLower(name)]
}

type sourceFile struct {
	path string
	size int64
}

type dedupMatch struct {
	source      sourceFile
	masterPaths []string
}

// Dedup finds files under the current directory (recursively, skipping its own
// "!found-in-master" subdirectory) that already exist somewhere under any of the
// given master dirs (compared by SHA256). Files smaller than minFileSize and
// files whose name is in ignoredNames (e.g. Thumbs.db) are skipped, and a match
// additionally requires the file extension to match
// (case-insensitive). A source file that shares a name and size with a master
// file but has a different checksum is reported as a possible bitrot and counted,
// but not moved. In write mode the matched (content-identical) files are moved
// into the "!found-in-master" subdirectory, preserving their relative path.
// Master dirs are read-only and never modified. This is a self-contained process:
// it uses no xattrs and requires no repo initialization.
func Dedup(masters []string, write bool) error {
	const source = "."

	if err := EnsureDir(source); err != nil {
		return err
	}

	absSource, err := resolvePath(source)
	if err != nil {
		return err
	}

	// Resolve masters, dropping duplicates (a wildcard or symlinks can yield the
	// same directory twice) and guarding each against the source dir.
	var masterDirs []string
	seen := make(map[string]bool)
	for _, master := range masters {
		if err := EnsureDir(master); err != nil {
			return err
		}
		absMaster, err := resolvePath(master)
		if err != nil {
			return err
		}
		if absSource == absMaster {
			return &pathError{master, "Master dir must not be the source dir"}
		}
		if strings.HasPrefix(absSource, absMaster+string(os.PathSeparator)) {
			return &pathError{master, "Source dir must not be inside a master dir"}
		}
		if seen[absMaster] {
			continue
		}
		seen[absMaster] = true
		masterDirs = append(masterDirs, master)
	}

	mastersLabel := strings.Join(masterDirs, ", ")
	if write {
		fmt.Printf("Compare %s with: %s\n", source, mastersLabel)
	} else {
		fmt.Printf("Dry-run compare %s with: %s\n", source, mastersLabel)
	}

	sourceFiles, err := listSourceFiles(source)
	if err != nil {
		return err
	}
	if len(sourceFiles) == 0 {
		fmt.Printf("Source dir %s has no files, nothing to do\n", source)
		return nil
	}

	// Stage 1: index source files by size, then scan masters for size matches.
	sizeIndex := make(map[int64][]sourceFile)
	for _, sf := range sourceFiles {
		sizeIndex[sf.size] = append(sizeIndex[sf.size], sf)
	}

	masterBySize := make(map[int64][]string)
	potential := 0
	for _, master := range masterDirs {
		walkErr := filepath.Walk(master, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				fmt.Printf("\rMaster scanning (1st): %s (found %d potential matches)\033[K", path, potential)
				return nil
			}
			if f.Mode().IsRegular() {
				if _, ok := sizeIndex[f.Size()]; ok {
					masterBySize[f.Size()] = append(masterBySize[f.Size()], path)
					potential++
				}
			}
			return nil
		})
		if walkErr != nil {
			fmt.Print("\n")
			return walkErr
		}
	}
	fmt.Print("\n")

	// Stage 2: confirm matches by checksum. Master hashes are cached so files
	// sharing a size are hashed at most once.
	masterHashes := make(map[string]string)
	masterHash := func(path string) string {
		if h, ok := masterHashes[path]; ok {
			return h
		}
		h := fileSha256(path)
		masterHashes[path] = h
		return h
	}

	var matches []dedupMatch
	mismatchCount := 0
	for _, sf := range sourceFiles {
		candidates := masterBySize[sf.size]
		if len(candidates) == 0 {
			continue
		}
		srcHash := fileSha256(sf.path)
		if srcHash == "" {
			continue
		}
		srcBase := filepath.Base(sf.path)
		var matched []string
		var mismatched []string
		for _, mp := range candidates {
			sameName := filepath.Base(mp) == srcBase
			// Content-identical matches must share an extension; same-name
			// candidates always do.
			if !sameName && !sameExt(sf.path, mp) {
				continue
			}
			if masterHash(mp) == srcHash {
				matched = append(matched, mp)
			} else if sameName {
				// Same name and size but different content: a likely bitrot.
				mismatched = append(mismatched, mp)
			}
		}
		if len(matched) > 0 {
			for i, mp := range matched {
				if i == 0 {
					fmt.Printf("Found: %s <-> %s\n", sf.path, mp)
				} else {
					fmt.Printf("Found: %s <-> %s (duplicate in master)\n", sf.path, mp)
				}
			}
			matches = append(matches, dedupMatch{sf, matched})
		}
		if len(mismatched) > 0 {
			for _, mp := range mismatched {
				fmt.Printf("Found: %s <-> %s - checksum mismatch!\n", sf.path, mp)
			}
			mismatchCount++
		}
	}

	// Stage 3: move matched source files into !found-in-master (write mode only).
	var movedCount int
	var movedBytes int64
	if write && len(matches) > 0 {
		foundDir := filepath.Join(source, "!found-in-master")
		for _, m := range matches {
			rel, err := filepath.Rel(source, m.source.path)
			if err != nil {
				fmt.Printf("Skipped by error: %s: %s\n", m.source.path, err)
				continue
			}
			dst := filepath.Join(foundDir, rel)
			if _, err := os.Stat(dst); err == nil {
				fmt.Printf("Skipped (target exists): %s\n", dst)
				continue
			}
			if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
				fmt.Printf("Skipped by error: %s: %s\n", m.source.path, err)
				continue
			}
			if err := os.Rename(m.source.path, dst); err != nil {
				fmt.Printf("Skipped by error: %s: %s\n", m.source.path, err)
				continue
			}
			movedCount++
			movedBytes += m.source.size
		}
	}

	if write {
		fmt.Printf("Moved %d files, deduplicated %s\n", movedCount, humanizeBytes(movedBytes))
	} else {
		var reclaim int64
		for _, m := range matches {
			reclaim += m.source.size
		}
		fmt.Printf("%d files can be moved, %s can be reclaimed\n", len(matches), humanizeBytes(reclaim))
	}
	if mismatchCount > 0 {
		fmt.Printf("%d files with checksum mismatch (possible bitrot)\n", mismatchCount)
	}

	return nil
}

// resolvePath returns the absolute, symlink-resolved path, falling back to the
// absolute path if symlink resolution fails.
func resolvePath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil
	}
	return real, nil
}

// listSourceFiles returns the regular files under dir (recursively), skipping
// dir's own top-level "!found-in-master" subdirectory and non-regular entries.
// Indexing progress is shown in place.
func listSourceFiles(dir string) (files []sourceFile, err error) {
	skip := filepath.Join(dir, "!found-in-master")
	err = filepath.Walk(dir, func(path string, f os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if f.IsDir() {
			if path == skip {
				return filepath.SkipDir
			}
			fmt.Printf("\rSource indexing: %s (indexed %d files)\033[K", path, len(files))
			return nil
		}
		if f.Mode().IsRegular() && f.Size() >= minFileSize && !isIgnored(f.Name()) {
			files = append(files, sourceFile{path: path, size: f.Size()})
		}
		return nil
	})
	fmt.Print("\n")
	return
}

// sameExt reports whether two paths have the same file extension, ignoring case.
func sameExt(a, b string) bool {
	return strings.EqualFold(filepath.Ext(a), filepath.Ext(b))
}
