package wpry

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseTheme(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fixture string
		want    Theme
	}{
		{
			"wp-org/1",
			"theme-wp-org-1.css",
			Theme{
				Name:        "Fabled Sunset",
				URI:         "https://example.com/fabled-sunset",
				Description: "Custom theme description...",
				Version:     "1.0.0",
				Author:      "Your Name",
				AuthorURI:   "https://example.com",
				Tags:        "block-patterns, full-site-editing",
				TextDomain:  "fabled-sunset",
				DomainPath:  "/assets/lang",
				TestedUpTo:  "6.4",
				RequiresWP:  "6.2",
				RequiresPHP: "7.4",
				License:     "GNU General Public License v2.0 or later",
				LicenseURI:  "https://www.gnu.org/licenses/gpl-2.0.html",
			},
		},
		{
			"full",
			"theme-full.css",
			Theme{
				Name:        "Full Theme",
				URI:         "https://example.com/full-theme",
				Description: "A fully specified theme.",
				Version:     "3.0.0",
				RequiresWP:  "6.0",
				RequiresPHP: "8.0",
				Author:      "Full Author",
				AuthorURI:   "https://example.com/author",
				License:     "GPL-2.0-or-later",
				LicenseURI:  "https://www.gnu.org/licenses/gpl-2.0.html",
				TextDomain:  "full-theme",
				DomainPath:  "/lang",
				TestedUpTo:  "6.5",
				Tags:        "custom-background, custom-logo",
				Template:    "twentytwentyfive",
			},
		},
		{
			"cr-only",
			"theme-cr-only.css",
			Theme{
				Name:    "CR Theme",
				Version: "0.2",
			},
		},
		{
			"crlf",
			"theme-crlf.css",
			Theme{
				Name:    "CRLF Theme",
				Version: "0.2",
			},
		},
		{
			"large-file",
			"theme-large-file.css",
			Theme{
				Name:    "Large File Theme",
				Version: "1.0",
			},
		},
		{
			"multiline",
			"theme-multiline.css",
			Theme{
				Name:    "MultiTheme",
				Version: "1.1",
			},
		},
		{
			"prefixed",
			"theme-prefixed.css",
			Theme{
				Name:    "PrefTheme",
				Version: "4.0",
			},
		},
		{
			"inline-comment",
			"theme-inline.css",
			Theme{
				Name:    "Inline Theme",
				Version: "2.0",
			},
		},
		{
			"child-theme",
			"theme-child.css",
			Theme{
				Name:     "Child Theme",
				Template: "twentytwentyfour",
				Version:  "1.0",
			},
		},
		{
			"block-open-star",
			"theme-block-open-star.css",
			Theme{
				Name:    "Block Open Star",
				Version: "2.1",
			},
		},
		{
			"block-open-star-close",
			"theme-block-open-star-close.css",
			Theme{
				Name:    "Block Open Star Close",
				Version: "3.1",
			},
		},
		{
			"block-star-close",
			"theme-block-star-close.css",
			Theme{
				Name:    "Block Star Close",
				Version: "4.1",
			},
		},
		{
			"block-open",
			"theme-block-open.css",
			Theme{
				Name:    "Block Open",
				Version: "6.1",
			},
		},
		{
			"block-open-close",
			"theme-block-open-close.css",
			Theme{
				Name:    "Block Open Close",
				Version: "7.1",
			},
		},
		{
			"block-close",
			"theme-block-close.css",
			Theme{
				Name:    "Block Close",
				Version: "8.1",
			},
		},
		{
			"star-slash-mid",
			"theme-star-slash-mid.css",
			Theme{
				Name:    "Mid Star",
				Version: "1.0",
			},
		},
		{
			"sandwich",
			"theme-sandwich.css",
			Theme{
				Name:    "Sandwich",
				Version: "3.0.0",
			},
		},
		{
			"duplicated",
			"theme-duplicated.css",
			Theme{
				Name:    "Duplicated",
				Version: "1.2.3",
			},
		},
	}

	for _, tt := range cases {
		for cs := range encoders() {
			name := fmt.Sprintf("%s/%s", tt.name, cs)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				f := openTestdata(t, tt.fixture)
				f = encodeReader(t, f, cs)

				got, err := ParseTheme(f)
				if err != nil {
					t.Fatalf("ParseTheme() unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("ParseTheme() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	}
}

func TestParseTheme_UTF8BOM(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		bom  []byte
	}{
		{
			"with-bom",
			[]byte{0xEF, 0xBB, 0xBF},
		},
		{
			"without-bom",
			[]byte{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := Theme{Name: "Foo"}

			b := slices.Concat(tt.bom, []byte("/* Theme Name: Foo */"))
			r := bytes.NewReader(b)

			got, err := ParseTheme(r)
			if err != nil {
				t.Fatalf("ParseTheme() unexpected error: %v", err)
			}
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ParseTheme() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseTheme_Error(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fixture string
	}{
		{
			"no-header",
			"theme-no-header.css",
		},
		{
			"beyond-8kb",
			"theme-beyond-8kb.css",
		},
		{
			"empty",
			"theme-empty.css",
		},
	}

	for _, tt := range cases {
		for cs := range encoders() {
			name := fmt.Sprintf("%s/%s", tt.name, cs)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				f := openTestdata(t, tt.fixture)
				f = encodeReader(t, f, cs)

				_, err := ParseTheme(f)

				want := errNoHeader
				if err == nil {
					t.Fatalf("ParseTheme() unexpected success, want err %v", want)
				}
				if !errors.Is(err, want) {
					t.Errorf("ParseTheme() error = %v, want %v", err, want)
				}
			})
		}
	}
}

func TestParseTheme_UnknownEncoding(t *testing.T) {
	t.Parallel()

	// Input that is invalid UTF-8 and unlikely to be decoded into valid
	// UTF-8 by the fallback heuristics. Parser should continue and return
	// errNoHeader rather than an encoding-specific error.
	b := []byte{0x00, 0x80, 0x81, 0xFF}
	r := bytes.NewReader(b)

	_, err := ParseTheme(r)
	if err == nil {
		t.Fatalf("ParseTheme() unexpected success, want err %v", errNoHeader)
	}
	if !errors.Is(err, errNoHeader) {
		t.Fatalf("ParseTheme() error = %v, want %v", err, errNoHeader)
	}
}
