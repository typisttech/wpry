## Why

The CLI currently reports successful results under a `path` field, and regular
file inputs can carry the full user-supplied path into that value. That makes
the JSON contract ambiguous for callers and conflicts with the CLI's intended
behavior of exposing only the matched file name.

## What Changes

- **BREAKING** Rename the CLI success JSON field from `path` to `file`.
- Ensure `file` contains only the basename of the parsed file for regular-file
  inputs, while directory and zip inputs continue to report the matched file
  name.
- Update the CLI spec delta and end-to-end CLI fixtures to assert the renamed
  field and basename-only behavior.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `cli`: Change the success JSON contract from `path` to `file` and require the
  CLI to emit only the basename of the matched file.

## Impact

- `cmd/wpry/main.go` success result shaping for file, directory, and zip
  outputs.
- `cmd/wpry/testdata/script/*.txt` fixture expectations and any related CLI
  coverage.
- Existing CLI consumers that read the `path` field from successful JSON
  output.
