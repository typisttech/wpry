## Why

The CLI currently accepts files and directories but does not handle zip archives. Users commonly distribute plugins and themes as zip files; allowing the CLI to accept zip files and parse them directly improves ergonomics and enables CI workflows that operate on packaged artifacts.

## What Changes

- Extend the `cli` capability so that when the `<path>` argument is a zip file, the CLI will open the archive and attempt to parse it as either a plugin or a theme by invoking existing FS-based helpers (`ParsePluginFS`, `ParseThemeFS`). This uses `archive/zip` from the Go standard library.
- Keep current behavior for files and directories unchanged. The zip-handling is an additive feature.

## Capabilities

### Modified Capabilities
- `cli`: The existing `cli` capability is extended to treat zip files as a first-class input type. No new capability is introduced; this is a modification of the existing `cli` behavior. The change is additive and non-breaking — the external CLI surface remains the same.

## Impact

- Affected code: `cmd/wpry/main.go` (argument dispatch). Instead of a custom adapter, the CLI will use `archive/zip` to open the archive and then use `github.com/spf13/afero/zipfs` (via `afero`) to obtain an in-memory filesystem. Wrap the resulting `afero.Fs` with `afero.NewIOFS` to obtain an `io/fs.FS` suitable for `ParsePluginFS` / `ParseThemeFS`.
- Dependencies: adds `github.com/spf13/afero` (v1.15.0) as a non-test dependency. NOTE: this repository's linter (`depguard`) currently restricts non-test third-party imports to the standard library, `golang.org/x/*`, and `github.com/typisttech/wpry`. Adding `afero` to non-test code may require a depguard rule update or an explicit exception. Options:
  - Request a depguard exception for `github.com/spf13/afero` (preferred if acceptable)
  - Vendor the small adapter layer (not preferred per current direction)
  - Keep a small custom adapter guarded behind a build tag as a last resort
- Tests: add testscript tests that generate zip archives at runtime using the `zip` CLI (testscript script runs `zip` to assemble plugin.zip / theme.zip), and unit tests for any helper code. If CI does not provide the `zip` binary, consider checking in small fixture archives as fallback.

- Error format: when a `.zip` file is supplied but cannot be opened or parsed, the CLI SHALL return the error JSON shape with the message prefixed by `invalid zip:`, for example: `{ "error":"invalid zip: <err>" }` and exit with code `1`.
