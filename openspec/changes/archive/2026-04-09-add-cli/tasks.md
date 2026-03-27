## 1. Setup & Dependencies

- [x] 1.1 Create testscript test directory `cmd/wpry/testdata/script/` and add a README describing how the tests are organized
- [x] 1.2 Add test-only dependency `github.com/rogpeppe/go-internal/testscript` to `go.mod` (or `go test` toolchain) so testscript tests can run in CI

## 2. CLI Implementation

- [x] 2.1 Implement path-type detection in `cmd/wpry/run` (use `os.Stat` or equivalent) and wire `cfg.timeout` into a context used for parsing
- [x] 2.2 Implement directory parsing flow: concurrently invoke `ParsePluginFS(ctx, fsys, WithMaxWorkers(cfg.parallel))` and `ParseThemeFS(ctx, fsys)` and return the first successful result; cancel other goroutines on success
- [x] 2.3 Implement file parsing flow (Option B): open the input file separately for each parser and run `ParsePlugin` and `ParseTheme` concurrently; the first successful parse wins and other readers/goroutines are cancelled/closed
- [x] 2.4 Produce a single compact JSON object on stdout for success (`{ "path": "...", "plugin": {...} }` or `{ "path": "...", "theme": {...} }`) and for errors produce `{ "error": "..." }`. All human-facing usage/help/diagnostics must go to stderr
- [x] 2.5 Ensure exit codes: `0` success, `2` invalid args (existing `mustParseInput` already handles this), `1` all other failures

## 3. JSON Marshaling & Struct Tags

- [x] 3.1 Add `json` struct tags (snake_case, `omitempty`) to `Plugin` fields to control JSON output
- [x] 3.2 Add `json` struct tags (snake_case, `omitempty`) to `Theme` fields to control JSON output

## 4. Tests (testscript)

- [x] 4.1 Add testscript script `invalid-args` that runs the built binary with no positional args and asserts exit code `2` and that usage text appears on stderr
- [x] 4.2 Add testscript script `plugin-file` that builds the binary, runs it against a known plugin fixture, asserts exit code `0`, unmarshals stdout JSON and verifies `path` and `plugin.name`/`plugin.version` fields
- [x] 4.3 Add testscript script `theme-file` that builds the binary, runs it against a `style.css` fixture, asserts exit code `0`, unmarshals stdout JSON and verifies `path` and `theme.name`/`theme.version`
- [x] 4.4 Add testscript script `directory` that builds the binary, runs it against a directory containing parseable candidates and accepts either a `plugin` or `theme` JSON result (do NOT assert which parser won)
- [x] 4.5 Add testscript script `error-case` that runs the binary against a directory or file with no headers and asserts exit code `1` and that stdout contains an `error` JSON object

## 5. Verification & CI

- [x] 6.1 Run `go test ./...` and ensure all unit tests (including existing concurrency/cancellation tests) pass
