## 1. Implementation

- [x] 1.1 Update `convertToUTF8([]byte) []byte` in `encoding.go` to:
  - Return original bytes immediately if `utf8.Valid`
  - Strip UTF-8 BOM if present
  - Detect and decode UTF-32 BOMs (BE/LE) using `utf32.UTF32(...).NewDecoder()`
  - Detect and decode UTF-16 BOMs (BE/LE) using `unicode.UTF16(...).NewDecoder()`
  - If no BOM and input not valid UTF-8, attempt decoding with `charmap.Windows1252` then `charmap.ISO8859_1`
  - Accept a decoded result only if it's valid UTF-8 and passes the existing "bad rune" sanity check
  - If all decoders fail, return the original bytes unchanged

- [x] 1.2 Keep `convertToUTF8` internal and do not change ParsePlugin/ParseTheme signatures

## 2. Tests

- [x] 2.1 Add/ensure unit tests in `encoding_test.go` cover:
  - UTF-8 (no-op)
  - UTF-16BE/LE with BOM
  - UTF-32BE/LE with BOM
  - Windows-1252
  - ISO-8859-1

- [x] 2.2 Add targeted integration tests in `plugin_test.go` and `theme_test.go` that iterate over encoders and assert parsed header values match expected strings

- [x] 2.3 Add edge-case tests:
  - Invalid or unknown encodings — ensure original bytes are returned and parser behaves predictably (likely errNoHeader)
  - Inputs with UTF-8 BOM — ensure BOM is stripped and content parsed

## 3. Verification

- [x] 3.1 Run `go test ./...` and fix any regressions

## 4. Documentation

- [x] 4.1 Add brief comment in `encoding.go` explaining heuristics and rationale (BOM-first, then charmap fallbacks). Do not expose implementation details in package documentation.
