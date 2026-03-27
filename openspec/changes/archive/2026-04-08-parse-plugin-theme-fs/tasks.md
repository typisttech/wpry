## 1. Setup

- [x] 1.1 Create helper function `ParsePluginFS` in the
  existing package with signatures matching repo conventions and accepting
  `context.Context` and functional options (including `WithMaxWorkers`).
- [x] 1.2 Create helper function`ParseThemeFS` in the
  existing package with signatures matching repo conventions and accepting
  `context.Context`.
- [x] 1.3 Add functional options type and `WithMaxWorkers(n int)` option with 
  default fallback to `runtime.GOMAXPROCS(0)` when `n <= 0`.
- [x] 1.4 Add any required imports and update module files (if needed).

## 2. Core Implementation

- [x] 2.1 Implement top-level candidate discovery: use `fs.ReadDir(fsys, ".")`
  and consider only regular files; for plugins match `.php` files
  (case-insensitive), for themes only `style.css` (case-insensitive).
- [x] 2.2 For each candidate, spawn worker goroutines (bounded by max workers)
  to attempt parsing using existing `ParsePlugin` / `ParseTheme` helpers.
- [x] 2.3 Implement fastest-first selection: return the first successful parse
  result and the relative path within the FS.
- [x] 2.4 Ensure per-file parse errors (e.g., `errNoHeader`) are ignored and
  do not become part of aggregated errors.

## 3. Concurrency and Cancellation

- [x] 3.1 Create a child context that can be cancelled when a worker returns
  success; workers must observe `ctx.Done()` and exit promptly.
- [x] 3.2 Ensure all opened files are closed in worker goroutines even when
  canceled or on error.
- [x] 3.3 Ensure no goroutine leaks: workers should return when parent/child
  context is cancelled.

## 4. Error Policy

- [x] 4.1 Return an error when no candidate yields a successful
  parse (do not export the error type).
- [x] 4.2 When caller's context is cancelled before success, return the
  context error (e.g., `context.Canceled`).

## 5. Tests

- [x] 5.1 Add table-driven tests for `ParsePluginFS` covering:
    - candidate enumeration (only `.php` files considered)
    - successful parse returns parsed Plugin and relative path
    - fastest-first selection semantics
    - all-negative case returns unexported error
- [x] 5.2 Add table-driven tests for `ParseThemeFS` covering:
    - `style.css` present and parsed successfully
    - `style.css` absent returns unexported error
    - cancellation behavior when `style.css` read blocks

## 6. Documentation / Cleanup

- [x] 6.1 Run `go test ./...` and fix issues discovered by the tests.
