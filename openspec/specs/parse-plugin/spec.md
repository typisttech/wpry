# parse-plugin Specification

## Purpose

Parse WordPress plugin header data from a file-like reader.

## Requirements

### Requirement: Read file prefix

The system SHALL read up to the first 8192 bytes (8 KiB) of the provided io.Reader when parsing plugin headers.

#### Scenario: Read limit enforced

- **WHEN** ParsePlugin is given a reader containing plugin data larger than 8 KiB with the header in the first 8 KiB
- **THEN** the parser SHALL find the header and return the parsed values

#### Scenario: Header beyond limit

- **WHEN** ParsePlugin is given a reader whose header appears after the first 8 KiB
- **THEN** the parser SHALL return an errNoHeader (header is considered missing)

### Requirement: Normalize line endings

The parser SHALL replace CR characters (`\r`) with LF (`\n`) before scanning for headers.

#### Scenario: CR-only line endings

- **WHEN** the input contains CR-only line endings
- **THEN** ParsePlugin SHALL normalize them and successfully parse headers present in the first 8 KiB

#### Scenario: CRLF line endings

- **WHEN** the input contains CRLF line endings
- **THEN** ParsePlugin SHALL normalize them and successfully parse headers present in the first 8 KiB

### Requirement: Per-header single-line matching

The parser SHALL locate header values using a regular expression equivalent to WordPress' get_file_data pattern. For a header named `Header Name` the effective regexp SHALL be (case-insensitive, multi-line):

`(?mi)^(?:[ \t]*<\?php)?[ \t\/*#@]*Header Name:(.*)$`

The parser SHALL capture only group 1 as the header value and SHALL NOT accept multi-line header values (fields do not span multiple lines).

#### Scenario: Header line matched

- **WHEN** a header line like `/* Plugin Name: My Plugin` appears in the first 8 KiB
- **THEN** ParsePlugin SHALL capture `My Plugin` as the Plugin Name value

#### Scenario: Multi-line values ignored

- **WHEN** a header value continues on subsequent lines (no colon on continuation lines)
- **THEN** ParsePlugin SHALL NOT append continuation lines; only the first matching line is used

### Requirement: Cleanup captured values

The parser SHALL call a cleanup routine equivalent to WordPress' _cleanup_header_comment on the captured value. The cleanup SHALL trim surrounding whitespace and remove trailing comment terminators such as `*/` or PHP close `?>` that may appear at the end of the captured value.

#### Scenario: Trailing comment terminator

- **WHEN** the captured value contains `*/` or `?>` on the same line
- **THEN** ParsePlugin SHALL remove the terminator and all content following it, then trim the value

### Requirement: Canonical headers and mapping

The parser SHALL recognize the canonical WordPress plugin headers (case-insensitive):

- Plugin Name (maps to Plugin.Name)
- Plugin URI (maps to Plugin.URI)
- Version
- Description
- Author
- Author URI
- License
- License URI
- Text Domain
- Domain Path
- Network
- Requires at least (maps to Plugin.RequiresWP)
- Requires PHP
- Update URI
- Requires Plugins

For each recognized header, ParsePlugin SHALL set the corresponding Plugin struct field.

### Requirement: Error handling

The parser SHALL return `errNoHeader` if the required name header is not found in the scanned region. If the input encoding is invalid UTF-8 and cannot be converted to valid UTF-8 by the internal conversion helper, the parser SHALL continue parsing the scanned region and MAY return `errNoHeader` if the mandatory name header is not found.

#### Scenario: Missing name

- **WHEN** the header block contains no `Plugin Name` entry
- **THEN** `ParsePlugin` SHALL return `errNoHeader`
