## Context

We currently accept files and directories as the CLI `<path>` argument and dispatch to the appropriate parsers:
- Single file: `ParsePlugin` / `ParseTheme`
- Directory: concurrently run `ParsePluginFS` and `ParseThemeFS`, returning the first success

We want to add support for zip files (typical distribution format for plugins/themes) without changing existing behavior.

## Goals / Non-Goals

**Goals:**
- Allow users to pass a zip archive to the CLI and have the tool attempt to parse it as a plugin or theme.
- Reuse existing FS-based parsing helpers (`ParsePluginFS`, `ParseThemeFS`) by exposing the zip archive as an `fs.FS`.
- Keep changes minimal and non-breaking.

**Non-Goals:**
- Implement arbitrary archive formats (tar, rar). Only zip is required.

## Decisions

- Use `archive/zip` from the standard library to open zip files, and use the community package `github.com/spf13/afero/zipfs` (v1.15.0) to treat the archive as an afero `Fs`. Convert the `afero.Fs` to the standard `fs.FS` using `afero.NewIOFS` so existing `ParsePluginFS` / `ParseThemeFS` can be reused without modification.
- Add a small helper in `cmd/wpry` that given a zip file path returns an `fs.FS` and a cleanup callback if necessary. Keep the helper small and testable. The helper responsibilities:
  - Open the zip archive with `zip.OpenReader`.
  - Create an `afero.Fs` via `zipfs.New(&reader)` and then call `afero.NewIOFS(aferoFs)` to obtain an `fs.FS`.
  - Detect and auto-strip a single top-level directory inside the archive by scanning file names; if present, return `fs.Sub` of the IOFS root to that prefix.
  - Return a cleanup function that closes the underlying `zip.ReadCloser`.
- CLI runtime flow: detect if `<path>` is a zip file (by file extension and `os.Stat` + basic header sniff if desired). If zip, open it and call `ParsePluginFS(ctx, fsys, WithMaxWorkers(cfg.parallel))` and `ParseThemeFS(ctx, fsys)` concurrently, mirroring directory behavior.

## Risks / Trade-offs

- Relying on file extension alone (extension-only detection) is the chosen approach. If the extension is `.zip`, attempt to open using `zip.OpenReader`. If opening fails, return a JSON error with the message `invalid zip: <err>` and exit code `1`.
- Memory/IO: large zip files could be expensive. `archive/zip` reads central directory into memory; acceptable for typical plugin/theme packages.
