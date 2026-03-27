## Why

Some plugin and theme files are encoded in legacy or multi-byte encodings (UTF-16 LE/BE, UTF-32, Windows-1252, ISO-8859-1). The current parsing helpers assume UTF-8 and either fail or return incorrect header values when presented with these encodings. Converting common encodings to UTF-8 inside the parsing pipeline increases robustness when scanning third-party WordPress code without introducing a heavy-weight charset detector.

## What Changes

- Add encoding-aware conversion inside the internal helper `convertToUTF8` so it can detect BOMs for UTF-16/UTF-32 (BE/LE) and perform conversions. For other single-byte legacy encodings (Windows-1252, ISO-8859-1) attempt decoding using `golang.org/x/text/encoding` decoders and fall back if a decoder fails.
- Use `encoding.Decoder` instances (from `golang.org/x/text/encoding`) to perform conversions and to enable fallback probing when the initial decoder doesn't produce valid UTF-8.
- Add unit tests that exercise conversion from UTF-16 LE/BE, UTF-32 LE/BE, Windows-1252, and ISO-8859-1 into UTF-8 and verify `ParsePlugin` / `ParseTheme` behave correctly with such inputs.
- Keep the public APIs (ParsePlugin, ParseTheme) unchanged; make the encoding logic internal to the package.

## Capabilities

### New Capabilities
- `encoding`: Convert common encodings to UTF-8 and expose only an internal helper used by parsing functions.

### Modified Capabilities
- `parse`: `ParsePlugin` and `ParseTheme` will transparently accept files encoded in common legacy encodings and UTF-16/32 by internally converting to UTF-8 before parsing. No change to the public API surface is required; behavior will be strictly additive.

## Impact

- Code: modify `encoding.go` to implement the improved `convertToUTF8` behaviour; add `golang.org/x/text/encoding` as a test or build dependency where needed.
- Tests: add `encoding_test.go` cases covering the listed encodings and update `plugin_test.go` and `theme_test.go` to validate parsing works when input is encoded in these encodings.
- Dependencies: introduce minimal dependency on `golang.org/x/text/encoding` (and specific subpackages such as `charmap` / `unicode` packages) to perform decoding. Keep usage internal to avoid API surface changes.
- Performance: conversions will only run when input is not valid UTF-8 or when a BOM indicates a specific UTF variant; normal UTF-8 inputs are processed as before.

## Constraints / Non-Goals

- Do not implement a comprehensive charset detector (e.g., Mozilla's Universal Charset Detector). Use deterministic heuristics: BOM detection for UTF-16/32 and simple decoder probing for known single-byte encodings.
- Do not expose encoding detection or decoders to callers — keep the logic internal.

## Next Steps (implementation tasks)

1. Implement `convertToUTF8([]byte) []byte` to:
   - Return the original bytes immediately if they are valid UTF-8.
   - Detect BOMs for UTF-16 and UTF-32 (LE/BE) and decode accordingly.
   - If no BOM and not valid UTF-8, attempt decoding with `charmap.Windows1252` and `charmap.ISO8859_1` decoders (using `encoding.Decoder`). If a decoder produces valid UTF-8, return the decoded bytes.
   - If all decoders fail, return the original content unchanged.

2. Add tests:
   - Unit tests for `convertToUTF8` with sample byte sequences for each encoding.
   - Integration tests where `ParsePlugin` / `ParseTheme` are given readers containing headers encoded in the above encodings and assertions that parsing succeeds and yields expected values.

3. Run `go test ./...` and adjust as necessary.

4. Document internal behavior in package comments if needed.
