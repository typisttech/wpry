## Context

Repository currently declares Plugin and Theme types and placeholder ParsePlugin/ParseTheme functions, but no implementation. The goal is to implement header parsing that mirrors WordPress' get_file_data/_cleanup_header_comment behavior for the first 8 KiB of a file. The parsing will be used by tools that inspect plugin PHP files and theme style.css files.

Constraints
- Prefer minimal standard-library-only implementation.
- Tests are required before implementation (per repo policy).
- Keep public API unchanged: ParsePlugin(io.Reader) and ParseTheme(io.Reader).

## Goals / Non-Goals

**Goals:**
- Implement ParsePlugin and ParseTheme to reliably extract all canonical header fields.
- Provide robust handling of common header comment formats (leading `*`, `/**` , `/*`, CSS `/*`).
- Add well-scoped unit tests covering canonical and edge cases.

**Non-Goals:**
- Support arbitrary non-UTF-8 encodings beyond validating UTF-8. We'll return a clear error for unsupported encodings rather than add an external dependency.
- Parse headers beyond the first 8 KiB of the file (matches WordPress behavior).

## Decisions

1. Read limit: Read up to 8192 bytes from the reader (WordPress convention). Rationale: avoids scanning large files and matches upstream behavior.

2. Encoding handling: Implement a simple convertToUTF8 that accepts valid UTF-8 input and returns an errUnsupportedEncoding for invalid sequences. Rationale: keeps implementation dependency-free and predictable. If later needed, we can add optional transcoding.

3. Header extraction semantics (match WordPress get_file_data):
- Read the first 8 KiB of the file and replace CR (`\r`) with LF (`\n`) to normalize line endings (WordPress does this before scanning).
- WordPress uses the following PCRE to locate a header on a single line (from get_file_data):

  /^(?:[ \t]*<\?php)?[ \t\/*#@]*' . preg_quote( $regex, '/' ) . ':(.*)$/mi

  where `$regex` is the header name (for example `Plugin Name` or `Theme Name`). Flags: `m` (multi-line) and `i` (case-insensitive).

- Practical Go equivalent: when searching for a header named `Header Name` compile a regexp like:

  `(?mi)^(?:[ \t]*<\?php)?[ \t\/*#@]*Header Name:(.*)$`

- This captures the remainder of the matching line after the first colon into group 1. Fields DO NOT span multiple lines; the value is taken from the first matching line only.
- Apply cleanupHeaderComment to the captured value to remove trailing comment terminators and trim whitespace.

4. Cleanup behavior: Implement a cleanupHeaderComment function that mirrors WordPress' _cleanup_header_comment semantics:
   - Strip trailing comment terminators (`*/`) or PHP close tags (`?>`) and anything that follows on that line
   - Trim surrounding whitespace from the captured header value
   - The cleanup function should be conservative and only remove trailing closers — it should not attempt to join multiple lines.

5. Header key matching: Use a pre-compiled `regexps` map keyed by lowercase identifiers (e.g., `"plugin_name"`, `"requires_at_least"`). `extractHeader` looks up the regex by key, runs `FindStringSubmatch`, and returns the cleaned captured value. All canonical headers map directly to Plugin/Theme struct fields.

6. Error conditions: Define errNoHeader and errUnsupportedEncoding. Return errNoHeader when no header value is found for required fields.

### Supported header names

The implementation will explicitly recognize the following WordPress header names (matching is case-insensitive but use the canonical labels below when storing in Extra):

Plugin headers (canonical label and struct field when different):

- Plugin Name (Name)
- Plugin URI (URI)
- Version
- Description
- Author
- Author URI
- License
- License URI
- Text Domain
- Domain Path
- Network
- Requires at least (RequiresWP)
- Requires PHP
- Update URI
- Requires Plugins

Theme headers (canonical label and struct field when different):

- Theme Name (Name)
- Theme URI (URI)
- Description
- Version
- Author
- Author URI
- Tags
- Text Domain
- Domain Path
- Tested up to
- Requires at least
- Requires PHP
- License
- License URI
- Template

## Risks / Trade-offs

- Risk: Some real-world plugin/theme files use exotic encodings. Mitigation: document behavior and add an optional transcoding step later if needed.
- Risk: Differences between WordPress' exact regex and our implementation could lead to small mismatches. Mitigation: write tests that mirror real-world header samples and iterate.

## Implementation notes

- Add small helper functions in the same package:
  - convertToUTF8([]byte) (string, error) — validate/convert input to UTF-8; returns errUnsupportedEncoding for invalid sequences
  - cleanupHeaderComment(string) string — remove comment terminators (`*/` or `?>`) and all following content on the same line, then trim surrounding whitespace; mirrors WordPress' _cleanup_header_comment
  - extractHeader(s, header string) string — look up the compiled regexp by lowercase key in the regexps map, run FindStringSubmatch, and return the cleaned captured value
  - regexps map[string]*regexp.Regexp — shared pre-compiled regexp map keyed by lowercase identifiers (e.g., "plugin_name", "requires_at_least")
- Keep changes localized to plugin.go and theme.go; tests in plugin_test.go and theme_test.go.
