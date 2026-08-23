# WPry

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
