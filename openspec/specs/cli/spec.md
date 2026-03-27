# Capability: cli

Purpose: Command-line interface for the wpry package. Defines the CLI's
behaviour, JSON output contract, exit codes, and test expectations.

## Requirements

### Requirement: Command-line arguments and exit codes
The CLI SHALL accept the flags `-parallel <n>` and `-timeout <d>` and a single
positional argument `<path>`.

The CLI SHALL exit with the following codes:
- `0` on success (parsed a plugin or theme)
- `2` for invalid arguments or usage errors (wrong positional count or empty path)
- `1` for all other failures (e.g., parse failures, I/O errors, timeouts)

#### Scenario: Invalid positional arguments
- **WHEN** the CLI is invoked with zero or more than one positional argument
- **THEN** the CLI SHALL write usage text to stderr and exit with code `2`

### Requirement: Path detection and parsing selection
The CLI SHALL detect whether `<path>` is a file, a directory, or a zip archive. Depending 
on the kind of path the CLI SHALL attempt parsing until the first successful result is 
available and then return that result (fastest-first semantics).

- If `<path>` is a file, the CLI SHALL attempt to parse it using the library
  helpers `ParsePlugin` and `ParseTheme` until one returns a successful
  parse; the CLI SHALL return the first successful parse that completes.
- If `<path>` is a directory, the CLI SHALL attempt to parse using `ParsePluginFS`
  and `ParseThemeFS` and return the first successful parse that completes.
- If `<path>` is a zip archive, the CLI SHALL open the archive and present it as
  an `fs.FS`; the CLI SHALL attempt to parse using `ParsePluginFS` and `ParseThemeFS`
  and return the first successful parse that completes.

When multiple concurrent parsing attempts are made, the CLI SHALL cancel other
ongoing attempts promptly once a successful parse is obtained so that no
goroutine or file handle leaks occur.

#### Scenario: File parses as plugin
- **WHEN** `<path>` names a regular file whose contents contain valid plugin
  headers within the required prefix
- **THEN** the CLI SHALL write a JSON object describing the plugin to stdout
  and exit with code `0`

#### Scenario: File parses as theme
- **WHEN** `<path>` names a regular file whose contents contain valid theme
  headers within the required prefix (for example, a standalone `style.css`)
- **THEN** the CLI SHALL write a JSON object describing the theme to stdout
  and exit with code `0`

#### Scenario: Directory fastest-first selection
- **WHEN** `<path>` names a directory containing multiple `.php` files where
  one candidate parses quickly and others are slow or blocking
- **THEN** the CLI SHALL return the result from the fastest successful
  candidate and promptly cancel other workers

#### Scenario: Directory parses as theme
- **WHEN** `<path>` names a directory containing a top-level `style.css` that
  parses successfully as a theme
- **THEN** the CLI SHALL write a JSON object describing the theme (with
  `path` set to `style.css`) to stdout and exit with code `0`

#### Scenario: Zip file containing a plugin
- **WHEN** the CLI is invoked with a path to a zip file whose contents include a valid plugin file
- **THEN** the CLI SHALL parse the plugin using `ParsePluginFS` and output the plugin JSON

#### Scenario: Zip file containing a theme
- **WHEN** the CLI is invoked with a path to a zip file whose contents include a valid theme (style.css)
- **THEN** the CLI SHALL parse the theme using `ParseThemeFS` and output the theme JSON

#### Scenario: Invalid zip file
- **WHEN** the CLI is invoked with a path to a file that is not a valid zip archive or cannot be opened
- **THEN** the CLI SHALL return a clear error explaining that the archive could not be opened or parsed

### Requirement: JSON output format
The CLI SHALL write exactly one JSON object to stdout on completion. The JSON
object SHALL be one of the following shapes:

- Success (plugin):
```
{
  "file": "<filename>",
  "plugin": { /* plugin fields, see below */ }
}
```

- Success (theme):
```
{
  "file": "<filename>",
  "theme": { /* theme fields, see below */ }
}
```

- Error:
```
{ "error": "<message>" }
```

For success objects, `file` SHALL be the file name (base name) containing the
headers (no directory component). The `plugin` and `theme` objects SHALL contain
the canonical fields used by the library (for plugins: `name`, `uri`,
`description`, `version`, `requires_wp`, `requires_php`, `author`, `author_uri`,
`license`, `license_uri`, `update_uri`, `text_domain`, `domain_path`,
`requires_plugins`, `network`; for themes: `name`, `uri`, `description`,
`version`, `requires_wp`, `requires_php`, `author`, `author_uri`, `license`,
`license_uri`, `text_domain`, `domain_path`, `tags`, `template`, `tested_up_to`).

Fields that are empty or unset SHALL be omitted from the JSON object (i.e.,
only non-empty fields are emitted).

#### Scenario: Plugin JSON shape
- **WHEN** the CLI parses a file that yields a Plugin with only `Name` and
  `Version` set
- **THEN** stdout SHALL contain a JSON object with `file` and `plugin` keys,
  and `plugin` SHALL only include `name` and `version` keys (other keys omitted)

#### Scenario: Theme JSON shape
- **WHEN** the CLI parses a file that yields a Theme with only `Name` and
  `Version` set
- **THEN** stdout SHALL contain a JSON object with `file` and `theme` keys,
  and `theme` SHALL only include `name` and `version` keys (other keys omitted)

#### Scenario: Error JSON and exit code
- **WHEN** the parser cannot produce a valid plugin or theme from `<path>` or
  an unexpected I/O/error occurs (excluding invalid-argument errors)
- **THEN** the CLI SHALL write `{ "error": "<message>" }` to stdout and
  exit with code `1`

### Requirement: Concurrency and configuration
The CLI SHALL respect the `-parallel` flag when invoking filesystem-based
parsers (`ParsePluginFS`, `ParseThemeFS`) by passing the value as the maximum
worker count (via `WithMaxWorkers`). If `-parallel` is 0 or negative the CLI
SHALL allow the parser to use `runtime.GOMAXPROCS(0)` as the default.

When parsing a regular file, the CLI SHALL open the file separately for each
parser (Option B) and run `ParsePlugin` and `ParseTheme` concurrently using
separate readers so they may run fastest-first without sharing a reader.

The CLI SHALL enforce `-timeout` as an overall timeout for parsing. If the
timeout elapses before any parser returns success, the CLI SHALL cancel
operations and exit with code `1` writing an error JSON to stdout.

#### Scenario: Timeout enforced
- **WHEN** the `-timeout` duration elapses before any parser completes
- **THEN** the CLI SHALL cancel parsing, write an error JSON to stdout and exit
  with code `1`

### Requirement: stdout/stderr separation
The CLI SHALL write only the single JSON object described above to stdout. Any
usage text, help, or diagnostic messages SHALL be written to stderr.

#### Scenario: No extra stdout output
- **WHEN** the CLI completes successfully
- **THEN** stdout SHALL contain exactly one JSON object and nothing else

### Requirement: Testscript-driven CLI tests
The repository SHALL include testscript tests that exercise the CLI binary for
the scenarios above. Tests SHALL build the `wpry` binary and run it with
fixtures to validate exit codes and JSON output shape. Concurrency behaviour
will not be asserted in testscript; that is covered by package unit tests.

#### Scenario: testscript exercises CLI
- **WHEN** testscript tests run the built `wpry` binary against prepared
  fixtures (files and directories)
- **THEN** tests SHALL be able to unmarshal stdout JSON and assert on its
  structure and values, and assert the binary's exit code. Testscript tests
  MUST NOT assert which parser wins in fastest-first scenarios — they should
  accept either a valid `plugin` or `theme` JSON result when both are possible.
