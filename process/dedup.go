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

// skipDirs are directory names dedup never descends into, on either the source
// or the master side. Matched by directory base name at any depth.
var skipDirs = map[string]bool{
	"@eaDir":           true, // Synology NAS index/thumbnail dir
	"!found-in-master": true, // dedup's own move target
}

// mediaExts are the file extensions treated as media when --media is set. Keys
// are lower-case and include the leading dot; matching is case-insensitive.
var mediaExts = map[string]bool{
	// images (incl. RAW)
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".heic": true,
	".heif": true, ".tiff": true, ".tif": true, ".webp": true, ".bmp": true,
	".cr2": true, ".nef": true, ".arw": true, ".dng": true, ".raf": true,
	".orf": true, ".rw2": true, ".pef": true,
	// video
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".m4v": true,
	".3gp": true, ".mts": true, ".m2ts": true, ".mpg": true, ".mpeg": true,
	".wmv": true, ".flv": true, ".webm": true,
	// audio
	".mp3": true, ".flac": true, ".wav": true, ".m4a": true, ".aac": true,
	".ogg": true, ".opus": true, ".wma": true, ".aiff": true,
}

func isMedia(name string) bool {
	return mediaExts[strings.ToLower(filepath.Ext(name))]
}

type sourceFile struct {
	path string
	size int64
}

type dedupMatch struct {
	source      sourceFile
	masterPaths []string
}

// Dedup finds files under the current directory (recursively, skipping any
// directory whose name is in skipDirs, e.g. "@eaDir" or its own
// "!found-in-master") that already exist somewhere under any of the given master
// dirs (compared by SHA256). Files smaller than minFileSize and files whose name
// is in ignoredNames (e.g. Thumbs.db) are skipped; when mediaOnly is set, only
// files with a media extension (see mediaExts) are considered. A match
// additionally requires the file extension to match (case-insensitive). A source
// file that shares a name and size with a master file but has a different checksum
// is reported as a possible bitrot and counted, but not moved. In write mode the
// matched (content-identical) files are moved into the "!found-in-master"
// subdirectory, preserving their relative path. Master dirs are read-only and
// never modified. This is a self-contained process: it uses no xattrs and requires
// no repo initialization.
func Dedup(masters []string, write, mediaOnly bool) error {
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

	sourceFiles, err := listSourceFiles(source, mediaOnly)
	if err != nil {
		return err
	}
	if len(sourceFiles) == 0 {
		fmt.Printf("Source dir %s has no files, nothing to do\n", source)
		return nil
	}

	// Stage 1: index source files by size.
	sourceSizeIndex := make(map[int64][]sourceFile)
	for _, sf := range sourceFiles {
		sourceSizeIndex[sf.size] = append(sourceSizeIndex[sf.size], sf)
	}

	// Stage 2: index master files by size, keeping only sizes present in the
	// source index (the implicit size semi-join that bounds memory for huge
	// masters).
	masterSizeIndex := make(map[int64][]string)
	potential := 0
	for _, master := range masterDirs {
		walkErr := filepath.Walk(master, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				if skipDirs[filepath.Base(path)] {
					return filepath.SkipDir
				}
				fmt.Printf("\rMaster indexing (stage 2): %s (found %d size matches)\033[K", path, potential)
				return nil
			}
			if f.Mode().IsRegular() {
				if _, ok := sourceSizeIndex[f.Size()]; ok {
					masterSizeIndex[f.Size()] = append(masterSizeIndex[f.Size()], path)
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

	// Stage 3: match each source file against its size candidates by
	// size → ext → hash. Master hashes are cached so files sharing a size are
	// hashed at most once.
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
		// Stage 3a: size matching — candidates that share this file's size.
		candidates := masterSizeIndex[sf.size]
		if len(candidates) == 0 {
			continue
		}

		// Stage 3b: ext matching (a future flag will make this optional).
		// Same name implies same ext, so bitrot candidates always survive here.
		srcBase := filepath.Base(sf.path)
		var compat []string
		for _, mp := range candidates {
			if !sameExt(sf.path, mp) {
				continue
			}
			compat = append(compat, mp)
		}
		if len(compat) == 0 {
			continue
		}

		// Stage 3c: hash matching — hash the source only now that a candidate
		// survived, then confirm by checksum.
		srcHash := fileSha256(sf.path)
		if srcHash == "" {
			continue
		}
		var matched []string
		var mismatched []string
		for _, mp := range compat {
			if masterHash(mp) == srcHash {
				matched = append(matched, mp)
			} else if filepath.Base(mp) == srcBase {
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

	// Stage 4: move matched source files into !found-in-master (write mode only).
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

// listSourceFiles returns the regular files under dir (recursively), skipping any
// directory whose name is in skipDirs (e.g. "@eaDir", "!found-in-master") and
// non-regular entries. When mediaOnly is set, only files with a media extension
// are returned. Indexing progress is shown in place.
func listSourceFiles(dir string, mediaOnly bool) (files []sourceFile, err error) {
	err = filepath.Walk(dir, func(path string, f os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if f.IsDir() {
			if skipDirs[filepath.Base(path)] {
				return filepath.SkipDir
			}
			fmt.Printf("\rSource indexing (stage 1): %s (indexed %d files)\033[K", path, len(files))
			return nil
		}
		if f.Mode().IsRegular() && f.Size() >= minFileSize && !isIgnored(f.Name()) && (!mediaOnly || isMedia(f.Name())) {
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
