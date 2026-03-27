## MODIFIED Requirements

### Requirement: Error handling

Both parsers SHALL return `errNoHeader` if the required name header is not found in the scanned region. If the input encoding is invalid UTF-8 and cannot be converted to valid UTF-8 by the internal conversion helper, the parser SHALL continue parsing the scanned region and MAY return `errNoHeader` if the mandatory name header is not found. The parser SHALL NOT surface an `errUnsupportedEncoding` error as a public API change — decoding failures are handled internally by returning the original bytes unchanged.

#### Scenario: Missing name — plugin (unchanged)
- **WHEN** the header block contains no `Plugin Name` entry
- **THEN** `ParsePlugin` SHALL return `errNoHeader`

#### Scenario: Missing name — theme (unchanged)
- **WHEN** the header block contains no `Theme Name` entry
- **THEN** `ParseTheme` SHALL return `errNoHeader`
