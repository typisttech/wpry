package wpry

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
)

// encoders returns a map of encoders used by tests.
func encoders() map[string]encoding.Encoding {
	return map[string]encoding.Encoding{
		"UTF-8":       encoding.Nop, // Original should be in UTF-8.
		"UTF-16BE":    unicode.UTF16(unicode.BigEndian, unicode.UseBOM),
		"UTF-16LE":    unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
		"UTF-32BE":    utf32.UTF32(utf32.BigEndian, utf32.UseBOM),
		"UTF-32LE":    utf32.UTF32(utf32.LittleEndian, utf32.UseBOM),
		"Windows1252": charmap.Windows1252,
		"ISO8859_1":   charmap.ISO8859_1,
	}
}

func encodeReader(t *testing.T, r io.Reader, charset string) io.Reader {
	t.Helper()

	enc, ok := encoders()[charset]
	if !ok {
		t.Fatalf("unknown charset %q", charset)
	}
	return transform.NewReader(r, enc.NewEncoder())
}

func TestConvertToUTF8(t *testing.T) {
	t.Parallel()

	for cs := range encoders() {
		t.Run(cs, func(t *testing.T) {
			t.Parallel()

			want := []byte("foo bar baz qux")
			r := encodeReader(t, bytes.NewReader(want), cs)
			b, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("io.ReadAll() unexpected error = %v", err)
			}

			got := convertToUTF8(b)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("convertToUTF8() = %q, want %q", got, want)
			}
		})
	}
}

func TestConvertToUTF8_UnrecognizedEncoding(t *testing.T) {
	t.Parallel()

	// Create an input that's invalid UTF-8 and unlikely to be decoded into
	// valid UTF-8 by Windows-1252 / ISO-8859-1. Use a short sequence of
	// control bytes including NUL and uncommon sequences.
	want := []byte{0x00, 0x80, 0x81, 0xFF}

	got := convertToUTF8(want)

	if !bytes.Equal(got, want) {
		t.Errorf("convertToUTF8() = %q, want %q", got, want)
	}
}

func TestTryDecode_RejectsReplacementChar(t *testing.T) {
	t.Parallel()

	// U+FFFD Replacement character sequence in UTF-8.
	content := []byte{0xff, 0xfe, 0xff}
	dec := encoding.Nop.NewDecoder()

	got, ok := tryDecode(dec, content)

	if ok {
		t.Fatalf("tryDecode() unexpected success, got: %q", got)
	}
}
