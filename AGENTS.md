# AGENTS.md

(generated)

Kuferek — a Go CLI to sync, compare, and deduplicate directory copies using
SHA256 checksums stored in each file's extended attributes (xattrs).

## Stack
- Go — module `kuferek`; `go.mod` says `go 1.13`, toolchain pinned to `1.20.7`
  in `.tool-versions`.
- `github.com/spf13/cobra` — CLI framework.
- `github.com/pkg/xattr` — extended-attribute storage for file/dir metadata.

## Build / Run / Test
- Build: `go build`  → produces `./kuferek`
- Run:   `./kuferek <command> [flags] [dir...]`
- Test:  `go test ./...`  (tests live in `process/`, e.g. `process/meta_test.go`)
- Format: `gofmt -w .`
- Cross-compile (MIPS LE): `./builld_mipsle.sh`

No Makefile, linter config, or CI.

## Layout
- `main.go` — entry point, calls `cmd.Start()`.
- `cmd/` — one cobra command per file: `root.go`, `init.go`, `scan.go`,
  `compare.go`, `merge.go`, `du.go`. CLI parsing only.
- `process/` — business logic, no cobra deps:
  - `directory.go` — repo init/validation (`InitRepo`, `EnsureRepo`).
  - `scanner.go` — recursive walk, hashing, directory comparison.
  - `meta.go` — read/write/parse per-file metadata in xattrs.
  - `file.go` — SHA256 hashing, size/time formatting.
  - `merge.go` — copy unique files, preserving relative structure.
  - `disk_usage.go` — real vs deduplicated usage stats.

## Commands
- `init <dir>...` — mark dir(s) as a repo (writes `user.kuferek-repo`; currently
  always `copy`).
- `scan <dir>...` — walk, compute SHA256, store metadata. `--verify` recomputes.
- `compare <dir1> <dir2>` — list files unique to each. `-1/--left` prints left-only
  (`>`), `-3/--right` prints right-only (`<`); pass neither and it prints nothing.
- `merge <dir> -t <target> [-m <master>]` — copy files in `<dir>` not present in
  master into target. `-o/--overwrite`, `-f/--force` (skip errors), `--verify`.
- `du <dir>...` — print file count, unique count, real and deduplicated bytes.

Global flags: `-m/--master` (default `.`), `-d/--debug`, `-e/--excludes`.

## Conventions & gotchas
- Metadata lives in xattrs, not a sidecar DB: per file `user.kuferek` =
  `"<sha256>,<size>,<modtime>"` plus `user.sha256`; per dir `user.kuferek-repo` =
  `master|copy`. A directory must be `init`ed before scan/compare/merge
  (`EnsureRepo` guards this).
- Naming: exported PascalCase = `process` package API; unexported camelCase =
  internal helpers; command vars are `cmd*`.
- Errors: `process`/`RunE` return `error`; custom `pathError`; `log.Fatal` on
  fatal CLI paths. Output via `fmt.Printf` / `log`.
- xattrs require Linux + a supporting filesystem (ext4/xfs/btrfs); they silently
  do not work on tmpfs, many NFS mounts, etc.
- `go.mod` (`go 1.13`) and `.tool-versions` (`1.20.7`) disagree — match existing
  Go style; don't assume newer language features are intended.
