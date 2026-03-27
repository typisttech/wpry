## Why

Parsing WordPress-style header comments from plugin PHP files and theme CSS files is currently unimplemented in this repository. Implementing robust header parsing unlocks reliable extraction of metadata (name, version, author, and arbitrary extra headers) for tools that analyze or package WordPress plugins and themes.

## What Changes

- Implement two parsing capabilities: `parse-plugin` and `parse-theme` that provide functions to parse header metadata from an io.Reader and return typed structs.
- Add canonical header mapping and a cleanup routine that mirrors WordPress behavior for header comment formats and continuations.
- Add tests that cover normal headers, star-prefixed comment lines, multiline/continuation values, and no-header / encoding error cases.

## Capabilities

### New Capabilities
- `parse-plugin`: Parse plugin PHP file header blocks into a Plugin struct (fields: name, uri, description, version, requires-wp, requires-php, author, author-uri, license, license-uri, update-uri, text-domain, domain-path, requires-plugins, extra).
- `parse-theme`: Parse theme CSS file header blocks (style.css) into a Theme struct (fields: name, uri, author, author-uri, description, version, requires-wp, tested-up-to, requires-php, license, license-uri, text-domain, domain-path, tags, template, extra).

### Modified Capabilities
- None

## Impact

- Code: plugin.go, theme.go (implementations of ParsePlugin and ParseTheme)
- Tests: add unit tests in plugin_test.go and theme_test.go covering canonical and edge cases
- No external dependencies planned; encoding handling will be conservative (validate UTF-8 and return an error for unsupported encodings)
