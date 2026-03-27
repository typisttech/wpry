# Capability: parse-theme-fs

Purpose: Parse WordPress theme header data from a filesystem.

### Requirement: Theme discovery limited to top-level style.css
The system SHALL only consider a top-level `style.css` file (case-insensitive
on the filename) as the theme candidate when parsing a theme from an
`fs.FS`. If `style.css` is absent at the top-level, the helper SHALL return
an error indicating the FS is not a theme.

#### Scenario: style.css present
- **WHEN** the FS contains `style.css` at the top-level with a valid theme
  header
- **THEN** `ParseThemeFS` returns the parsed `Theme` and path `style.css`.

#### Scenario: style.css absent
- **WHEN** the FS does not contain a top-level `style.css`
- **THEN** `ParseThemeFS` returns an error.

### Requirement: Parse theme headers
`ParseThemeFS` SHALL attempt to parse the theme header in `style.css` using
the existing `ParseTheme` function. On success it SHALL return the parsed
`Theme` and the relative path within the FS.

#### Scenario: Successful theme parse
- **WHEN** `style.css` contains a valid theme header
- **THEN** `ParseThemeFS` returns the parsed `Theme` and the path
  `style.css`.

### Requirement: Concurrency and cancellation
`ParseThemeFS` SHALL respect the same bounded concurrency and cancellation
semantics as `ParsePluginFS` (child context cancel on first success; workers
respect `ctx.Done()`). Although themes are single-file candidates in the
common case, tests SHOULD exercise cancellation semantics with a blocking
`style.css` file to ensure no leaks and timely shutdown.

#### Scenario: Cancellation for theme parsing
- **WHEN** `style.css` is readable but the read blocks, and the caller
  cancels the context
- **THEN** `ParseThemeFS` returns promptly with the caller's context error and
  the blocking read goroutine completes after observing cancellation.
