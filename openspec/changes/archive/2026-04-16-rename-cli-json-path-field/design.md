## Context

`cmd/wpry/main.go` uses a CLI-local `result` struct with a JSON tag of
`path`, and `runFile` currently stores the user-supplied file path directly in
that field. This means successful regular-file parses can emit an absolute or
relative input path, while directory and zip-backed parses already return only
the matched file name such as `plugin.php` or `style.css`.

The requested change is a contract update for the CLI only. The parsing
capabilities in the library already return the information the CLI needs, and
the current testscript fixtures under `cmd/wpry/testdata/script/` cover the
success cases that will need their JSON assertions updated.

## Goals / Non-Goals

**Goals:**
- Rename the CLI success JSON field from `path` to `file`.
- Ensure successful regular-file parses emit only the basename of the parsed
  file.
- Preserve existing parser selection, timeout, error handling, and plugin/theme
  payloads.

**Non-Goals:**
- Do not change the `wpry` library APIs or the path values returned by
  `ParsePluginFS` and `ParseThemeFS`.
- Do not add backward-compatible aliases such as emitting both `path` and
  `file`.
- Do not change the error JSON shape.

## Decisions

1. Rename the CLI-local result field to `File` and change its JSON tag to
   `json:"file,omitzero"`.
   Rationale: this keeps the contract change isolated to `cmd/wpry` without
   introducing translation logic in `render`.
   Alternative considered: map `path` to `file` only at render time. Rejected
   because the success payload is already modeled explicitly by the `result`
   struct and should carry the final contract directly.

2. Normalize only regular-file outputs with `filepath.Base(path)` inside
   `runFile`.
   Rationale: the bug is that file inputs reuse the full user-supplied path.
   Directory and zip flows already return matched file names from the parsing
   helpers, so they do not need extra rewriting.
   Alternative considered: apply `filepath.Base` to every successful result in
   `render`. Rejected because it would hide where path normalization happens and
   would couple output shaping to presentation instead of result creation.

3. Update the existing CLI script fixtures to assert `file` instead of `path`.
   Rationale: the current scripts already cover file, directory, and zip success
   cases affected by this contract change.
   Alternative considered: add a second layer of dedicated contract tests.
   Rejected because the existing end-to-end coverage is already the right place
   to assert the JSON shape.

## Risks / Trade-offs

- [Breaking JSON contract] Existing consumers reading `path` will fail.
  Mitigation: mark the change as breaking in the proposal/spec and scope the
  change narrowly to the documented success payload.
- [Missed fixture updates] Leaving one script on `path` would create mixed
  contract coverage.
  Mitigation: update every success fixture under `cmd/wpry/testdata/script/`
  that currently asserts `path`, then verify with `go test ./cmd/wpry`.
- [Over-normalization] Applying basename rewriting too broadly could mask future
  parser path semantics.
  Mitigation: perform basename normalization only in `runFile` and leave FS
  parser return values unchanged.
