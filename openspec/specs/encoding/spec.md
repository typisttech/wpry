# Capability: encoding

Purpose: Best-effort conversion of common legacy encodings to UTF-8 for internal parsing.

## Requirements

### Requirement: Convert common encodings to UTF-8

The system SHALL provide an internal helper `convertToUTF8([]byte) []byte` that attempts to convert input bytes to UTF-8 using deterministic heuristics. The helper SHALL:

- Return the original bytes immediately if they are valid UTF-8.
- Detect BOMs for UTF-16 and UTF-32 (both little- and big-endian) and decode accordingly using reliable decoders.
- If no BOM is present and the input is not valid UTF-8, attempt decoding using `charmap.Windows1252` and `charmap.ISO8859_1` in that order. If a decoder produces valid UTF-8 and passes a basic sanity check, return the decoded bytes.
- If all decoding attempts fail, return the original bytes unchanged.

#### Scenario: UTF-8 input preserved
- **WHEN** `convertToUTF8` is given bytes that are valid UTF-8
- **THEN** it SHALL return the original bytes unchanged

#### Scenario: UTF-16/32 BOM decoding
- **WHEN** `convertToUTF8` is given bytes beginning with a UTF-16 or UTF-32 BOM
- **THEN** it SHALL decode according to the BOM endianness and return valid UTF-8 bytes

#### Scenario: Single-byte legacy decoding
- **WHEN** `convertToUTF8` is given bytes that are not valid UTF-8 and have no BOM
- **THEN** it SHALL attempt `Windows-1252` then `ISO-8859-1` decoding and return the first decoded result that is valid UTF-8 and passes sanity checks

#### Scenario: Unrecognized encodings
- **WHEN** `convertToUTF8` cannot decode the input with any supported decoder
- **THEN** it SHALL return the original bytes unchanged

### Requirement: Internal-only behavior

The encoding conversion helper SHALL be internal to the package and NOT exposed to callers. Public functions (`ParsePlugin`, `ParseTheme`) SHALL remain with their current signatures and error semantics.
