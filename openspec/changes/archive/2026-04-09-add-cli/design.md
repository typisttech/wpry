## Context

The repository already exposes a tested library that parses WordPress plugin
and theme header metadata. There is a minimal CLI scaffold in `cmd/wpry/` but
the command currently only parses flags and returns. The goal is to implement
the CLI behaviour described in the proposal while reusing existing parsing
capabilities and tests.

Constraints and immediate considerations:
- Keep the wpry package API unchanged; the CLI must call into existing helpers.
- Tests will use `github.com/rogpeppe/go-internal/testscript` to run the built
  binary and assert on stdout/exit codes.
- The CLI is a single-command binary (no subcommands). Keep the UX simple and
  deterministic.

## Goals / Non-Goals

**Goals:**
- Implement argument parsing, path-type detection (file vs dir), invocation of
  ParsePlugin/ParseTheme/ParsePluginFS/ParseThemeFS as appropriate, and JSON
  rendering of results to stdout.
- Ensure graceful cancellation and bounded concurrency by honoring
  `-parallel` and `-timeout` flags and using context cancellation.
- Add end-to-end tests with testscript that validate JSON output, exit codes,
  and concurrency behaviour.

**Non-Goals:**
- Do not change the parsing behavior or requirements in existing package-level
  functions. Do not refactor library internals as part of this change.

## Decisions

1. Single-responsibility flow: the CLI parses flags and determines path type,
   then delegates to a small wrapper that exposes the minimal orchestration
   logic. This keeps main.go thin and easy to test.

2. Fastest-success selection: for directories, use ParsePluginFS and
   ParseThemeFS concurrently (bounded by `-parallel`) and return the first
   successful result. This mirrors the library semantics for FS helpers and
   keeps CLI behaviour predictable. The CLI will also apply fastest-first for
   single-file inputs by opening the file separately for each parser (Option B),
   allowing concurrent ParsePlugin and ParseTheme runs when parsing a regular
   file.

3. Output format: always write one JSON object to stdout. On success the
   object contains the parsed result; on error it contains `{ "error": "..." }`.
   The CLI exits 0 on success, 2 for invalid arguments, and 1 for other
   failures (matching the proposal and config.go usage).

4. Tests: use testscript to build the binary and run scenarios. Test scripts
   will be placed under `cmd/wpry/testdata/script/`, create temporary
   files/dirs, run `wpry` with flags, and assert on stdout/exit codes. Tests
   will perform JSON-aware assertions by unmarshalling stdout using a small
   `go` helper. Concurrency behaviour will not be asserted in testscript; the
   package unit tests already cover concurrency correctness. CLI tests will
   validate end-to-end behaviour (exit codes, JSON shape, and happy/error
   paths). Testscript tests MUST NOT assert which parser wins in
   fastest-first scenarios — they should accept either a valid `plugin` or
   `theme` JSON result when both are possible.

## Risks / Trade-offs

- Risk: Adding testscript tests introduces a test-only dependency. Mitigation:
  keep tests minimal and document the dependency in the change proposal.
- Trade-off: Concurrent parsing of single files (fastest-first) trades
  determinism for latency; you accepted nondeterministic fastest-first as the
  desired behavior. The implementation will open the file separately for each
  parser (Option B) to avoid sharing readers.
