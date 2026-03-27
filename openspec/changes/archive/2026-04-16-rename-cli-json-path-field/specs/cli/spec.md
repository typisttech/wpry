## MODIFIED Requirements

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
  `file` set to `style.css`) to stdout and exit with code `0`

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
headers (no directory component), even when the CLI input `<path>` names a file
using an absolute or relative path. The `plugin` and `theme` objects SHALL contain
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

#### Scenario: File input emits basename only
- **WHEN** the CLI is invoked with a regular file path that includes directory
  components and the file parses successfully
- **THEN** the success JSON SHALL set `file` to only the basename of the parsed
  file and SHALL NOT include any directory component from the user-supplied
  input path
