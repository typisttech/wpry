## Why

Provide a tiny, single-command CLI wrapper around the existing wpry library so
users can run the parser from the command line. This makes the library
accessible for one-off inspections, CI validation, and integration into scripts.

## What Changes

- Add a single-binary CLI command `wpry` that accepts a single positional
  argument `<path>` and flags `-parallel` and `-timeout`.
- The CLI will detect whether `<path>` is a file or directory and invoke the
  appropriate library helpers (`ParsePlugin`, `ParseTheme`, `ParsePluginFS`,
  `ParseThemeFS`) until one succeeds. The first successful parse is rendered as
  JSON to stdout. On error, the CLI emits a JSON error object to stdout and
  exits non-zero.
-- Add tests using `github.com/rogpeppe/go-internal/testscript` to exercise the
  CLI end-to-end, including exit codes and stdout JSON shape. Concurrency
  behaviour will not be asserted in testscript; concurrency correctness is
  covered by existing package unit tests.

- Implementation notes / choices:
  - For single-file inputs the CLI will run parsers concurrently (fastest-first)
    by opening the file separately for each parser. This mirrors the
    fastest-first behaviour used for FS helpers while avoiding a shared seek
    point.
  - JSON output will be compact (no pretty-printing). Error objects are written
    to stdout (human-facing usage text remains on stderr).
  - The CLI will not support `-` as a path for reading stdin.
  - The Plugin and Theme structs in the library will be updated with `json`
    struct tags (snake_case, `omitempty`) to align JSON output with the spec.

## Capabilities

### New Capabilities
- `cli`: Command-line interface for the wpry package. Covers argument parsing,
  invocation of library helpers based on path type, JSON output format, exit
  codes, and testscript-driven CLI tests.

### Modified Capabilities
- None. Existing parsing capabilities (parse-plugin, parse-theme,
  parse-plugin-fs, parse-theme-fs, encoding) are already implemented and
  tested; the CLI will *call* these capabilities but will not change their
  requirements.

## Impact

- New files under `cmd/wpry/` (already present) will be completed to implement
  the CLI behaviour and tested.
- New tests under `testdata` + `cmd` tests will use `testscript` to exercise
  the binary. This adds a test dependency on
  `github.com/rogpeppe/go-internal/testscript` (test-only). Testscript tests
  will be placed under `cmd/wpry/testdata/script/` and will perform JSON-aware
  assertions by unmarshalling stdout with a small `go` helper.

  These testscript tests MUST NOT assert which parser wins in fastest-first
  (nondeterministic) scenarios; they should accept either a valid `plugin` or
  `theme` JSON result when both are possible.

- Source change: the Plugin and Theme structs will gain `json` tags to control
  marshaling. This is not a breaking API change (function signatures remain the
  same) but it modifies the package's JSON encoding behavior.
