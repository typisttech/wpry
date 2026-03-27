## 1. CLI Output Contract

- [x] 1.1 Rename the CLI success JSON field in `cmd/wpry/main.go` from `path` to `file`
- [x] 1.2 Normalize regular-file success results to emit `filepath.Base(path)` while leaving directory and zip parser result paths unchanged

## 2. CLI Test Coverage

- [x] 2.1 Update the existing success script fixtures under `cmd/wpry/testdata/script/` to assert `file` instead of `path`
- [x] 2.2 Add or update coverage so a successful regular-file invocation verifies the JSON emits only the basename when the input path includes directories

## 3. Verification

- [x] 3.1 Run `go test ./cmd/wpry` and confirm the CLI contract change passes end-to-end tests
