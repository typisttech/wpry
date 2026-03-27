## ADDED Requirements

### Requirement: Discover top-level plugin candidates
The system SHALL enumerate the top-level entries of the provided `fs.FS` using
`fs.ReadDir(fsys, ".")` and consider only regular files whose name ends with
`.php` (case-insensitive) as plugin candidates. Subdirectories SHALL be
ignored.

#### Scenario: Candidate enumeration
- **WHEN** the FS contains `index.php`, `readme.txt`, and a subdirectory
  `lib/`
- **THEN** the discovery step considers only `index.php` as a candidate and
  ignores `readme.txt` and `lib/`.

### Requirement: Parse plugin headers
For each candidate file the system SHALL attempt to parse WordPress-style
plugin headers using the existing `ParsePlugin` function. A candidate that
returns a successful parse SHALL be considered a valid plugin result.

#### Scenario: Successful header parse
- **WHEN** `index.php` contains a valid plugin header
- **THEN** `ParsePluginFS` returns the parsed `Plugin` and the relative path
  `index.php`.

### Requirement: Selection semantics (fastest-first)
When multiple candidates are processed concurrently, `ParsePluginFS` SHALL
return the first successful parse that completes (fastest-first). Callers
SHALL NOT rely on deterministic selection order when multiple files are
valid.

#### Scenario: Fastest-first selection
- **WHEN** `fast.php` parses quickly and `slow.php` would also parse
  successfully but takes longer
- **THEN** `ParsePluginFS` returns the result from `fast.php` and does not
  wait for `slow.php` to finish.

### Requirement: Bounded concurrency and configuration
The implementation SHALL use a bounded worker pool. The default maximum
workers SHALL be `runtime.GOMAXPROCS(0)`. The API SHALL expose
`WithMaxWorkers(n int)` as a ParseOption. If `n` is 0 or negative, the
implementation SHALL fall back to the default.

#### Scenario: Max workers option
- **WHEN** `WithMaxWorkers(2)` is passed
- **THEN** `ParsePluginFS` uses at most two concurrent workers processing
  candidates.

### Requirement: Cancellation and cooperative shutdown
`ParsePluginFS` SHALL accept a `context.Context`. When a candidate parse
succeeds, `ParsePluginFS` SHALL cancel a child context so that all other
workers observe `ctx.Done()` and return promptly. If the caller's context is
cancelled before any success, `ParsePluginFS` SHALL return promptly with the
context error.

#### Scenario: Cancellation (fast-success cancels blocker)
- **WHEN** the FS has `fast.php` (valid header) and `block.php` (Open/Read
  blocks), and `ParsePluginFS` is invoked with `WithMaxWorkers(2)`
- **THEN** `ParsePluginFS` returns the parsed `Plugin` and path for
  `fast.php`, and the blocked worker observing `block.php` returns after
  observing `ctx.Done()` (no goroutine leak).

#### Scenario: Caller cancels before success
- **WHEN** the caller's context is cancelled before any candidate parses
  successfully
- **THEN** `ParsePluginFS` returns promptly with the caller's context error
  (e.g., `context.Canceled`).

### Requirement: Error policy
`ParsePluginFS` SHALL silently continue on individual-file parse errors such
as `errNoHeader` or unexpected I/O errors. Such per-file errors SHALL NOT be
aggregated or logged by the helper. If no candidate yields a successful
parse, `ParsePluginFS` SHALL return an error indicating no 
candidate succeeded.

#### Scenario: All-negative
- **WHEN** all candidate `.php` files lack valid headers or fail to parse
- **THEN** `ParsePluginFS` returns an error and does not leak
  goroutines.
