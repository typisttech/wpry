package wpry

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	xunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
)

// convertToUTF8 performs a best-effort conversion of the provided byte slice to
// UTF-8.
//
// Heuristics (in order):
//  1. If input is already valid UTF-8, return it (strip UTF-8 BOM if present)
//  2. Check for UTF-32 BOMs (BE/LE) and decode when present
//  3. Check for UTF-16 BOMs (BE/LE) and decode when present
//  4. Try Windows-1252
//  5. Try ISO-8859-1
//
// Only accept a decoded result if it's valid UTF-8 and passes a small set of
// sanity checks (no replacement chars, no NULs, no Unicode non-characters, and
// no unexpected control characters). If all attempts fail, return the original
// bytes unchanged.
func convertToUTF8(content []byte) []byte {
	// Strip UTF-8 BOM up front.
	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(content) {
		return content
	}

	encodings := []encoding.Encoding{
		utf32.UTF32(utf32.BigEndian, utf32.ExpectBOM),
		xunicode.UTF16(xunicode.BigEndian, xunicode.ExpectBOM),
		charmap.Windows1252,
		charmap.ISO8859_1,
	}

	for _, enc := range encodings {
		dec := enc.NewDecoder()
		out, ok := tryDecode(dec, content)
		if !ok {
			continue
		}
		return out
	}

	return content
}

// tryDecode applies the provided transformer to content and returns the
// transformed bytes if they are valid UTF-8 and pass the sanity checks.
func tryDecode(dec transform.Transformer, content []byte) ([]byte, bool) {
	transformed, _, err := transform.Bytes(dec, content)
	if err != nil {
		return nil, false
	}
	if !utf8.Valid(transformed) {
		return nil, false
	}

	// Quick check for UTF-8 replacement character sequence (U+FFFD).
	if bytes.Contains(transformed, []byte{0xEF, 0xBF, 0xBD}) {
		return nil, false
	}

	for i := 0; i < len(transformed); {
		r, sz := utf8.DecodeRune(transformed[i:])
		i += sz
		if r == utf8.RuneError || isBadRune(r) {
			return nil, false
		}
	}

	return transformed, true
}

// isBadRune reports whether r is a rune that indicates a decode failure or
// otherwise makes the decoded content unsuitable for parsing headers.
func isBadRune(r rune) bool {
	if r == 0 {
		return true
	}

	// Unicode non-characters: U+FDD0..U+FDEF
	if r >= 0xFDD0 && r <= 0xFDEF {
		return true
	}

	// Plane-end non-characters: any code point whose low 16 bits are 0xFFFE or 0xFFFF
	if (r&0xFFFF) == 0xFFFE || (r&0xFFFF) == 0xFFFF {
		return true
	}

	// Control characters (except allowed whitespace)
	if unicode.IsControl(r) && !unicode.IsSpace(r) {
		return true
	}

	return false
}
