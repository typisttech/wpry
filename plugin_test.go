package wpry

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

//go:embed testdata
var testdata embed.FS

func openTestdata(t *testing.T, path string) io.Reader {
	t.Helper()

	p := filepath.Join("testdata", path)
	f, err := testdata.Open(p)
	if err != nil {
		t.Fatalf("testdata.Open(%q) unexpected error: %v", p, err)
	}
	t.Cleanup(func() { _ = f.Close() })

	return f
}

func TestParsePlugin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fixture string
		want    Plugin
	}{
		{
			"wp-org/1",
			"plugin-wp-org-1.php",
			Plugin{
				Name:            "My Basics Plugin",
				URI:             "https://example.com/plugins/the-basics/",
				Description:     "Handle the basics with this plugin.",
				Version:         "1.10.3",
				RequiresWP:      "5.2",
				RequiresPHP:     "7.2",
				Author:          "John Smith",
				AuthorURI:       "https://author.example.com/",
				License:         "GPL v2 or later",
				LicenseURI:      "https://www.gnu.org/licenses/gpl-2.0.html",
				UpdateURI:       "https://example.com/my-plugin/",
				TextDomain:      "my-basics-plugin",
				DomainPath:      "/languages",
				RequiresPlugins: "my-plugin, yet-another-plugin",
			},
		},
		{
			"wp-org/2",
			"plugin-wp-org-2.php",
			Plugin{
				Name:            "Plugin Name",
				URI:             "https://example.com/plugin-name",
				Description:     "Description of the plugin.",
				Version:         "1.0.0",
				RequiresWP:      "5.2",
				RequiresPHP:     "7.2",
				Author:          "Your Name",
				AuthorURI:       "https://example.com",
				License:         "GPL v2 or later",
				LicenseURI:      "http://www.gnu.org/licenses/gpl-2.0.txt",
				UpdateURI:       "https://example.com/my-plugin/",
				TextDomain:      "plugin-slug",
				RequiresPlugins: "my-plugin, yet-another-plugin",
			},
		},
		{
			"cr-only",
			"plugin-cr-only.php",
			Plugin{
				Name:    "CR Plugin",
				Version: "1.0",
			},
		},
		{
			"crlf",
			"plugin-crlf.php",
			Plugin{
				Name:    "CRLF Plugin",
				Version: "1.0",
			},
		},
		{
			"large-file",
			"plugin-large-file.php",
			Plugin{
				Name:    "Large File Plugin",
				Version: "1.0",
			},
		},
		{
			"multiline",
			"plugin-multiline.php",
			Plugin{
				Name:    "Multi",
				Version: "2.0",
			},
		},
		{
			"prefixed",
			"plugin-prefixed.php",
			Plugin{
				Name:    "Prefixed",
				Version: "3.0",
			},
		},
		{
			"full",
			"plugin-full.php",
			Plugin{
				Name:            "Full Plugin",
				URI:             "https://example.com/full-plugin",
				Description:     "A fully specified plugin.",
				Version:         "2.0.0",
				RequiresWP:      "6.0",
				RequiresPHP:     "8.0",
				Author:          "Full Author",
				AuthorURI:       "https://example.com/author",
				License:         "GPL-2.0-or-later",
				LicenseURI:      "https://www.gnu.org/licenses/gpl-2.0.html",
				UpdateURI:       "https://example.com/update",
				TextDomain:      "full-plugin",
				DomainPath:      "/lang",
				RequiresPlugins: "woocommerce, akismet",
				Network:         "true",
			},
		},
		{
			"inline-comment",
			"plugin-inline.php",
			Plugin{
				Name:    "Inline Plugin",
				Version: "5.0",
			},
		},
		{
			"hash-prefix",
			"plugin-hash.php",
			Plugin{
				Name:    "Hash Plugin",
				Version: "6.0",
			},
		},
		{
			"block-open-star",
			"plugin-block-open-star.php",
			Plugin{
				Name:    "Block Open Star",
				Version: "2.1",
			},
		},
		{
			"block-open-star-close",
			"plugin-block-open-star-close.php",
			Plugin{
				Name:    "Block Open Star Close",
				Version: "3.1",
			},
		},
		{
			"block-star-close",
			"plugin-block-star-close.php",
			Plugin{
				Name:    "Block Star Close",
				Version: "4.1",
			},
		},
		{
			"block-open",
			"plugin-block-open.php",
			Plugin{
				Name:    "Block Open",
				Version: "6.1",
			},
		},
		{
			"block-open-close",
			"plugin-block-open-close.php",
			Plugin{
				Name:    "Block Open Close",
				Version: "7.1",
			},
		},
		{
			"block-close",
			"plugin-block-close.php",
			Plugin{
				Name:    "Block Close",
				Version: "8.1",
			},
		},
		{
			"double-slash",
			"plugin-double-slash.php",
			Plugin{
				Name:    "Double Slash",
				Version: "10.1",
			},
		},
		{
			"php-close",
			"plugin-php-close.php",
			Plugin{
				Name:    "PHP Close",
				Version: "9.0",
			},
		},
		{
			"star-slash-mid",
			"plugin-star-slash-mid.php",
			Plugin{
				Name:    "Mid Star",
				Version: "1.0",
			},
		},
		{
			"php-close-mid",
			"plugin-php-close-mid.php",
			Plugin{
				Name:    "Mid PHP Close",
				Version: "2.0",
			},
		},
		{
			"sandwich",
			"plugin-sandwich.php",
			Plugin{
				Name:    "Sandwich",
				Version: "3.0.0",
			},
		},
		{
			"duplicated",
			"plugin-duplicated.php",
			Plugin{
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

				got, err := ParsePlugin(f)
				if err != nil {
					t.Fatalf("ParsePlugin() unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("ParsePlugin() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	}
}

func TestParsePlugin_UTF8BOM(t *testing.T) {
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

			want := Plugin{Name: "Foo"}

			b := slices.Concat(tt.bom, []byte("<?php /* Plugin Name: Foo */"))
			r := bytes.NewReader(b)

			got, err := ParsePlugin(r)
			if err != nil {
				t.Fatalf("ParsePlugin() unexpected error: %v", err)
			}
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ParsePlugin() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParsePlugin_Error(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fixture string
	}{
		{
			"no-header",
			"plugin-no-header.php",
		},
		{
			"beyond-8kb",
			"plugin-beyond-8kb.php",
		},
		{
			"empty",
			"plugin-empty.php",
		},
	}

	for _, tt := range cases {
		for cs := range encoders() {
			name := fmt.Sprintf("%s/%s", tt.name, cs)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				f := openTestdata(t, tt.fixture)
				f = encodeReader(t, f, cs)

				_, err := ParsePlugin(f)

				want := errNoHeader
				if err == nil {
					t.Fatalf("ParsePlugin() unexpected success, want err %v", want)
				}
				if !errors.Is(err, want) {
					t.Errorf("ParsePlugin() error = %v, want %v", err, want)
				}
			})
		}
	}
}

func TestParsePlugin_UnknownEncoding(t *testing.T) {
	t.Parallel()

	// Input that is invalid UTF-8 and unlikely to be decoded into valid
	// UTF-8 by the fallback heuristics. Parser should continue and return
	// errNoHeader rather than an encoding-specific error.
	b := []byte{0x00, 0x80, 0x81, 0xFF}
	r := bytes.NewReader(b)

	_, err := ParsePlugin(r)
	if err == nil {
		t.Fatalf("ParsePlugin() unexpected success, want err %v", errNoHeader)
	}
	if !errors.Is(err, errNoHeader) {
		t.Fatalf("ParsePlugin() error = %v, want %v", err, errNoHeader)
	}
}
