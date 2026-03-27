## Context

We need filesystem-aware helpers that enumerate top-level files from an
`fs.FS` and attempt to parse WordPress-style header comments with existing
parsers. Callers require cooperative cancellation and a way to bound
concurrency for resource control.

## Goals / Non-Goals

**Goals:**
- Provide `ParsePluginFS` and `ParseThemeFS` helpers with a sensible default
  concurrency cap and a functional options API.
 - Use `fs.ReadDir(fsys, ".")` and only consider regular files in the top-level
   directory. For plugins, match files with a `.php` extension (case-insensitive).
   For themes, only a top-level `style.css` file (case-insensitive on the
   filename) will be considered. If `style.css` is absent the function SHALL
   return an error indicating the FS is not a theme. Tests should
   construct temporary directories to exercise discovery semantics.
- Return the discovered relative path within `fsys` for the successful parse.
- Be cancelable: when a parse succeeds, cancel other workers immediately; when
  the provided ctx is cancelled, return promptly.

**Non-Goals:**
- Recursive traversal of subdirectories.
- Changing the behavior of `ParsePlugin` and `ParseTheme` functions.

## Decisions

1. Selection semantics: fastest-first (return the first successful parse that
   completes). This minimizes latency and aligns with the caller's concurrency
   expectation.
2. Error handling: continue-on-error (ignore `errNoHeader` and other unexpected
   parse or I/O errors for individual files). If no file yields a successful
   parse, return an error indicating no candidate succeeded.
3. Concurrency: expose a functional option `WithMaxWorkers(n int)` and use
   default `GOMAXPROCS`. If the caller passes 0 or a negative value, the 
   implementation SHALL fall back to the default.
4. Cancellation: create a child context and call cancel when a success is
   observed; workers must respect ctx.Done() and ensure files are closed. Tests
   SHOULD use a small custom in-test fs to reliably simulate blocking Open/Read
   behavior and assert no goroutine leaks.

## Risks / Trade-offs

- Trade-off: fastest-first may produce non-deterministic selection when multiple
  files are valid; tests must avoid relying on ordering.

## Scenario: Cancellation

- Purpose: Verify ParsePluginFS/ParseThemeFS cancel in-flight work after a
  successful parse and return promptly if the caller cancels.
- Setup:
  1. Create an in-test fs implementing fs.FS. The fs should expose at least two
     candidate files:
     - `fast.php` / `style.css`: returns a valid header quickly when opened.
     - `block.php`: simulates a slow or blocking Open/Read. Opening or Read
       will block until the test signals.
  2. Run ParsePluginFS/ParseThemeFS with a small worker cap via
     `WithMaxWorkers(2)` to ensure multiple workers run concurrently.
- Assertions:
  - When the fast candidate produces a successful parse, the function returns
    (parsed struct, path, nil) where `path` matches the fast candidate.
  - After the function returns, the blocked candidate's goroutine completes
    (the test should either unblock it or otherwise verify no goroutine leak).
  - The blocked worker should observe `ctx.Done()` and return without
    performing further work.
- Additional cases to include in tests:
  - Caller-cancel: cancel the caller's context before any parse completes;
    the function should return promptly with `ctx.Err()`.
  - All-negative: if all candidate files lack headers, the function should
    return an error and not leak goroutines.
- Test notes:
  - Tests MUST be table-driven per repo conventions.
  - Use `t.Cleanup` to cancel contexts and close any test channels.
  - To detect goroutine leaks: capture `runtime.NumGoroutine` before starting
    the call and poll (with timeout) for it to return to baseline after the
    call returns. Allow a small slack (e.g., +2) for transient goroutines.
  - Prefer `os.DirFS` + temp dirs for simple positive/negative cases; reserve
    a custom blocking FS only for the cancellation scenario.

Notes on test FS implementation (guidance, not code):
- Implement a test-only FS that returns a file object whose `Read` blocks on
  a channel. The test controls the channel and may unblock only after the
  function under test returns, verifying cooperative cancellation. Ensure the
  file's `Close` unblocks `Read` so workers can exit promptly when canceled.
