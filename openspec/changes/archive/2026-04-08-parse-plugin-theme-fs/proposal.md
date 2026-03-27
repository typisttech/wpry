## Why

Tools that analyze WordPress plugins and themes need a robust, cancellable
way to discover and parse metadata from a provided filesystem (os.FS). The
existing parsers operate on an io.Reader; callers that work with directories
want helper APIs that enumerate top-level files concurrently, respect
context cancellation, and return the first successful parse result.

This change provides file-system-aware discovery helpers for plugins and
themes with conservative concurrency and cooperative cancellation.

## What Changes

- Add two new filesystem discovery functions with configurable concurrency:
  - `ParsePluginFS(ctx context.Context, fsys fs.FS, opts ...ParseOption) (Plugin, string, error)`
  - `ParseThemeFS(ctx context.Context, fsys fs.FS) (Theme, string, error)`
- Implement functional options with a helper `WithMaxWorkers(n int) ParseOption`.
- Implement discovery logic in `plugin_fs.go` and `theme_fs.go` at the repo root
  (package `wpry`).
- Add unit tests that exercise cancellation and negative cases. Tests SHOULD
  create temporary directories (via testing helper or os.DirFS) to simulate
  filesystems rather than relying on flat fixtures in `testdata/`.

## Capabilities

### New Capabilities
- `parse-plugin-fs`: Discover plugin PHP files at the top-level of an `fs.FS`,
  parse headers concurrently, and return the first successful `Plugin` with the
  discovered relative file path string.
- `parse-theme-fs`: Discover theme CSS files (style.css) at the top-level of an
  `fs.FS`, parse headers concurrently, and return the first successful `Theme`
  with the discovered relative file path string.

### Modified Capabilities
- None

## Impact

- Code: add `plugin_fs.go` and `theme_fs.go` in package `wpry`.
  - Implement bounded worker pool (default `GOMAXPROCS`).
  - Use `fs.ReadDir(fsys, ".")` and consider only regular files. For plugins,
    match files with a `.php` extension (case-insensitive). For themes, only
    a top-level `style.css` file (case-insensitive on the filename) will be
    considered; if `style.css` is absent, the function SHALL return an
    unexported `errNotTheme`.
  - On a successful parse, cancel the context for other workers and return the
    parsed struct and the relative path (within `fsys`) exactly as returned by
    the FS.
  - Error policy: silently continue on `errNoHeader` or unexpected I/O/parse
    errors for individual files; do not aggregate or log these errors.

- Tests: add unit tests in `plugin_fs_test.go` and `theme_fs_test.go`.
  - Tests MUST construct temporary directories (via testing helper or os.DirFS)
    and place a `style.css` file for theme tests; theme discovery is strict and
    only `style.css` will be parsed.
  - Include cancellation test(s); do not assert strict first-success ordering.
