## Context

The `best-effort-encoding` change aims to make the package's header parsing resilient to common legacy and multi-byte encodings (UTF-16 LE/BE, UTF-32 LE/BE, Windows-1252, ISO-8859-1) without changing public APIs. Currently ParsePlugin and ParseTheme assume UTF-8 input and call an internal helper `convertToUTF8` before normalizing line endings and scanning the first 8 KiB. The proposal already specifies the behavioral requirements and tests.

Constraints and important context (do not copy into implementation files):
- Keep public API (ParsePlugin, ParseTheme) unchanged.
- Use deterministic heuristics (BOM detection for UTF-16/32; simple probing for single-byte legacy encodings).
- Prefer a fast-path for valid UTF-8 inputs to avoid perf regressions.
- Keep encoding logic internal to the package; do not expose decoders to callers.

## Goals / Non-Goals

**Goals:**
- Implement a best-effort `convertToUTF8([]byte) []byte` that: returns input unchanged if already valid UTF-8; detects and decodes UTF-16/32 BOM-marked content; probes Windows-1252 and ISO-8859-1 decoders when no BOM and input is not valid UTF-8; returns original bytes if decoding fails.
- Ensure ParsePlugin and ParseTheme behavior is unchanged for UTF-8 inputs and strictly additive for other encodings.
- Add unit and integration tests covering UTF-8, UTF-16 BE/LE, UTF-32 BE/LE, Windows-1252, and ISO-8859-1.

**Non-Goals:**
- Do not implement a general-purpose charset detector.
- Do not change public function signatures or error semantics (avoid adding error returns for encoding detection).

## Decisions

1. Fast-path using utf8.Valid
- Rationale: Most files are UTF-8; checking `utf8.Valid` first avoids unnecessary decoding overhead. This preserves performance for the common case.

2. BOM-first detection for UTF-16/UTF-32
- Rationale: BOMs reliably indicate endianness for UTF-16/32. Use `golang.org/x/text/encoding/unicode` and `utf32` decoders for BOM-aware decoding. Handle UTF-32 BOMs explicitly since UTF-32 isn't covered by unicode.UTF16.

3. Fallback probing with single-byte encodings
- Rationale: After ruling out valid UTF-8 and BOMs, attempt `charmap.Windows1252` and `charmap.ISO8859_1`. Accept a decoded result only if it is valid UTF-8 and passes a small heuristic (no excessive non-graphic control runes).

4. Keep convertToUTF8 signature as-is
- Rationale: Changing the function to return an error would force changes to read/ParsePlugin/ParseTheme and the public behavior. The spec does mention `errUnsupportedEncoding` but making it actionable would be a breaking behavior change. Prefer returning original bytes and letting parsing continue (existing errNoHeader behavior remains).

5. Use x/text transform.Bytes with NewDecoder
- Rationale: Simple and robust API for decoding whole byte slices. Implementation will use `transform.Bytes(enc.NewDecoder(), data)`.

## Risks / Trade-offs

[Risk] False positives from single-byte decoders → [Mitigation] Validate decoded bytes with `utf8.Valid` and a small "bad rune" filter (reject results containing many non-graphic control runes).

[Risk] Collisions where BOM-less UTF-16 is misinterpreted by single-byte decoders → [Mitigation] Prefer BOM detection first; single-byte decoders are only tried when input is not valid UTF-8 and no BOM detected.

[Risk] Tests rely on BOM generation/encoding helpers → [Mitigation] Use existing encoding_test.go helpers (encoders map) and add focused tests for BOM variants.

## Migration / Deployment Plan

1. Implement convertToUTF8 in `encoding.go` per the decisions above.
2. Run tests: `go test ./...`. Fix any failing tests.
3. Add new unit tests for edge cases (invalid encoding fallback) and targeted integration tests for ParsePlugin/ParseTheme with BOM variants.
4. Because this change is internal and additive, no special migration steps or rollbacks are required. If unintended behavior is observed, revert the conversion changes and investigate test failures.

## Open Questions

- Should we ever surface a dedicated `errUnsupportedEncoding` error when no decoding is possible? Current design keeps the API unchanged and returns original bytes instead.
- Are there other single-byte encodings commonly used in plugins/themes that we should include (e.g., Windows-1251) or would that be premature?

## Implementation Checklist (developer-friendly)

- [ ] Update `convertToUTF8([]byte) []byte` in `encoding.go`:
  - [ ] Early return when `utf8.Valid`.
  - [ ] Strip UTF-8 BOM if present and validate.
  - [ ] Detect UTF-32 BOMs and decode with `utf32.UTF32(...).NewDecoder()`.
  - [ ] Detect UTF-16 BOMs and decode with `unicode.UTF16(...).NewDecoder()`.
  - [ ] Attempt `charmap.Windows1252` and `charmap.ISO8859_1` decoding as fallbacks.
  - [ ] On success return decoded bytes, otherwise return original content.
- [ ] Add unit tests in `encoding_test.go` and integration cases in `plugin_test.go` and `theme_test.go` as specified in the proposal.
- [ ] Run full test suite and adjust heuristics if necessary.

---

This design.md was created based on the existing proposal and repository layout. It keeps the change small, internal, and test-driven, and lists clear next steps and risks. Once you confirm, I will create the `convertToUTF8` implementation and any tests you want me to add.
