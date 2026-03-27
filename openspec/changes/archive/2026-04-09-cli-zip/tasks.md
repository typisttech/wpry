## 1. CLI argument handling

- [x] 1.1 Detect when the CLI `<path>` argument is a zip file. Use file extension and attempt to open with `archive/zip` as confirmation.
- [x] 1.2 Add a helper to open a zip file and expose it as an `fs.FS` suitable for `ParsePluginFS` / `ParseThemeFS`.

  Notes: Use `github.com/spf13/afero/zipfs` (v1.15.0) to obtain an `afero.Fs`, then convert to `fs.FS` with `afero.NewIOFS`. Remember the repository's linter `depguard` may block adding `afero` as a non-test dependency; coordinate a depguard change if needed.

## 2. Parsing flow

- [x] 2.1 For zip input, mirror the directory behavior: concurrently invoke `ParsePluginFS(ctx, fsys, WithMaxWorkers(cfg.parallel))` and `ParseThemeFS(ctx, fsys)` and return the first successful result. Cancel the other goroutine on success.
- [x] 2.2 Ensure errors from opening the zip are surfaced with a helpful message and do not silently fallback to other handlers.

## 3. Tests

- [x] 3.1 Add testscript tests covering:
  - zip file containing a plugin; CLI should parse as plugin and print JSON
  - zip file containing a theme; CLI should parse as theme and print JSON
  - invalid zip file: CLI should fail with clear error
  - Notes: The testscript scripts will generate zip archives at runtime using the `zip` CLI. If CI lacks `zip`, add fixture zips as a fallback.
  - zip file containing a plugin; CLI should parse as plugin and print JSON
  - zip file containing a theme; CLI should parse as theme and print JSON
  - invalid zip file: CLI should fail with clear error
- [x] 3.2 Add unit tests for the zip->fs helper if introduced.
