## 1. Tests & Fixtures

 - [x] 1.1 Add golden fixtures for typical plugin header (plugin-basic) and theme header (theme-basic)
 - [x] 1.2 Add edge-case fixtures: plugin-no-header, theme-no-header, plugin-encoding-bad
 - [x] 1.3 Implement unit tests in plugin_test.go covering: happy path, missing name -> errNoHeader, encoding error -> errUnsupportedEncoding
 - [x] 1.4 Implement unit tests in theme_test.go covering: happy path, missing name -> errNoHeader, encoding error -> errUnsupportedEncoding

Note: Place all test fixtures / golden files under the repository-level `testdata/` directory. Use the following naming conventions for fixtures:

- Plugin fixtures: `testdata/plugin-<name>.php` (e.g., `testdata/plugin-full.php`, `testdata/plugin-no-header.php`)
- Theme fixtures: `testdata/theme-<name>.css` (e.g., `testdata/theme-full.css`, `testdata/theme-no-header.css`)

## 2. Helpers

 - [x] 2.1 Implement convertToUTF8([]byte) (string, error) that validates UTF-8 and returns errUnsupportedEncoding on invalid input
 - [x] 2.2 Implement cleanupHeaderComment(string) string to remove comment terminators (`*/` or `?>`) and all following content on the same line, then trim surrounding whitespace (match WordPress semantics)
 - [x] 2.3 Implement extractHeader(s, header string) string that applies the PCRE-equivalent regexp and returns the trimmed captured value
 - [x] 2.4 Add a shared pre-compiled regexps map keyed by lowercase identifiers (e.g., "plugin_name", "requires_at_least"), mapping each canonical header name to its compiled *regexp.Regexp

## 3. Core Parsing Implementation

 - [x] 3.1 Implement ParsePlugin(io.Reader) using the helpers, respecting the 8 KiB read limit, CR normalization, and regex matching per header
 - [x] 3.2 Implement ParseTheme(io.Reader) similarly for theme headers
 - [x] 3.3 Ensure ParsePlugin/ParseTheme return errNoHeader when name field is missing after parsing

## 4. QA, Lint, and Documentation

 - [x] 4.1 Run `go test ./...` and ensure all tests pass
 - [x] 4.2 Run `golangci-lint run` and fix any issues
 - [x] 4.3 Update README or package docs if necessary to document ParsePlugin/ParseTheme behavior
