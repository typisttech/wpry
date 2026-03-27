# WPry

## Toolchain
- `go.mod` pins Go `1.26.2`. The code intentionally uses newer stdlib/testing features (`sync.WaitGroup.Go`, `testing/synctest`, `os.OpenRoot`, `json:",omitzero"`); do not treat those as compatibility bugs.
- `mise.toml` pins `golangci-lint` `2.11`; run lint as `golangci-lint run` and formatting as `golangci-lint fmt`.

## Layout
- Root package `github.com/typisttech/wpry` is the library: header parsing, encoding normalization, and `fs.FS` helpers.
- `cmd/wpry` is the CLI. It races plugin and theme parsing for files, directories, and zip archives, then returns the first successful result. If both parses could succeed, the winner is intentionally nondeterministic.

## Verification
- Full suite: `go test ./...`
- Library-only loop: `go test .`
- CLI/testscript loop: `go test ./cmd/wpry`
- Refresh `testscript` golden files only when output intentionally changed: `WPRY_UPDATE_SCRIPTS=1 go test ./cmd/wpry`
- For non-trivial changes, also run `golangci-lint run`

## Gotchas
- CLI script tests live in `cmd/wpry/testdata/script/*.txt` and the zip scenarios shell out to `zip`.
- Preserve fixture line endings in `testdata/`; `.editorconfig` forces CR and CRLF for specific parser fixtures.
- `.golangci.yml` uses strict `depguard`: non-test code is limited to stdlib, `golang.org/x`, `github.com/spf13/afero`, and this module. Tests must not add `testify`; use stdlib or `github.com/google/go-cmp/cmp`.

## Specs
- If behavior is unclear, consult current specs in `openspec/specs/*/spec.md`. Files under `openspec/changes/archive/` are historical context, not the source of truth.
