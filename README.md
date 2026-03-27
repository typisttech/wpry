# WPry

`wpry` parses WordPress plugin and theme headers.

It ships as:
- a Go library for reading metadata from files and `fs.FS`
- a CLI that emits JSON for a file, directory, or zip archive

## Features
- Parse WordPress plugin headers from PHP files.
- Parse WordPress theme headers from `style.css`.
- Parse unzipped plugin and theme directories.
- Parse zip archives containing a plugin or theme.
- Normalize CR and CRLF line endings like WordPress.
- Apply best-effort encoding fallback before header parsing.

## Install
Build the CLI from source:

```bash
go build ./cmd/wpry
```

Or run it without installing:

```bash
go run ./cmd/wpry <path>
```

## CLI Usage

```bash
wpry [flags] <path>
```

Flags:
- `-parallel <n>`: maximum workers for filesystem parsing. `0` or less uses `GOMAXPROCS`.
- `-timeout <d>`: overall parse timeout. Default `1m`.

`<path>` may point to:
- a plugin PHP file
- a theme stylesheet
- an unzipped plugin directory
- an unzipped theme directory
- a zip archive containing either layout

When a path could parse as both a plugin and a theme, the CLI returns the first successful result. That selection is intentionally nondeterministic.

## Examples
Parse a plugin file:

```bash
go run ./cmd/wpry testdata/plugin-full.php
```

```json
{"path":"testdata/plugin-full.php","plugin":{"name":"Full Plugin","uri":"https://example.com/full-plugin","description":"A fully specified plugin.","version":"2.0.0","requires_wp":"6.0","requires_php":"8.0","author":"Full Author","author_uri":"https://example.com/author","license":"GPL-2.0-or-later","license_uri":"https://www.gnu.org/licenses/gpl-2.0.html","update_uri":"https://example.com/update","text_domain":"full-plugin","domain_path":"/lang","requires_plugins":"woocommerce, akismet","network":"true"}}
```

Parse a theme file:

```bash
go run ./cmd/wpry testdata/theme-full.css
```

```json
{"path":"testdata/theme-full.css","theme":{"name":"Full Theme","uri":"https://example.com/full-theme","author":"Full Author","author_uri":"https://example.com/author","description":"A fully specified theme.","version":"3.0.0","requires_wp":"6.0","tested_up_to":"6.5","requires_php":"8.0","license":"GPL-2.0-or-later","license_uri":"https://www.gnu.org/licenses/gpl-2.0.html","text_domain":"full-theme","domain_path":"/lang","tags":"custom-background, custom-logo","template":"twentytwentyfive"}}
```

Parse a directory:

```bash
go run ./cmd/wpry /path/to/wp-content/plugins/woocommerce
go run ./cmd/wpry /path/to/wp-content/themes/twentytwentynine
```

Parse a zip archive:

```bash
go run ./cmd/wpry /path/to/plugin.zip
go run ./cmd/wpry /path/to/theme.zip
```

## JSON Output
Successful output is exactly one JSON object on stdout.

Plugin result:

```json
{
  "path": "plugin.php",
  "plugin": {
    "name": "Full Plugin",
    "version": "2.0.0"
  }
}
```

Theme result:

```json
{
  "path": "style.css",
  "theme": {
    "name": "Full Theme",
    "version": "3.0.0"
  }
}
```

Error result:

```json
{
  "error": "no header found"
}
```

Exit codes:
- `0`: success
- `1`: parse, I/O, zip, or timeout failure
- `2`: invalid CLI usage

## Library Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/typisttech/wpry"
)

func main() {
	f, err := os.Open("plugin.php")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p, err := wpry.ParsePlugin(f)
	if err != nil {
		panic(err)
	}

	fmt.Println(p.Name)
}
```

Available entrypoints:
- `wpry.ParsePlugin(io.Reader)`
- `wpry.ParseTheme(io.Reader)`
- `wpry.ParsePluginFS(context.Context, fs.FS, ...ParseOption)`
- `wpry.ParseThemeFS(context.Context, fs.FS)`

## Parsing Behavior
- Only the first 8 KiB of each file is scanned for headers.
- CR is normalized to LF before header matching.
- UTF-8 BOM is stripped.
- If input is not valid UTF-8, decoding falls back to BOM-detected UTF-16/UTF-32, then Windows-1252, then ISO-8859-1.
- Header matching is intentionally aligned with WordPress header parsing semantics.

## Development
Run tests:

```bash
go test ./...
```

Run only library tests:

```bash
go test .
```

Run only CLI script tests:

```bash
go test ./cmd/wpry
```

Refresh CLI golden files only when output intentionally changed:

```bash
WPRY_UPDATE_SCRIPTS=1 go test ./cmd/wpry
```

Run lint and formatting:

```bash
golangci-lint run
golangci-lint fmt
```

Development notes:
- CLI script tests live in `cmd/wpry/testdata/script/*.txt`.
- Zip testscript scenarios shell out to `zip`.
- Preserve fixture line endings in `testdata/`; some fixtures intentionally use CR or CRLF.
- `.golangci.yml` enforces strict dependency rules. Tests use stdlib assertions or `github.com/google/go-cmp/cmp`.
