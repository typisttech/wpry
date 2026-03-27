package wpry

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var errNoHeader = errors.New("no header found")

// Plugin represents parsed [WordPress plugin headers].
//
// [WordPress plugin headers]: https://developer.wordpress.org/plugins/plugin-basics/header-requirements/
type Plugin struct {
	Name            string `json:"name,omitempty"`
	URI             string `json:"uri,omitempty"`
	Description     string `json:"description,omitempty"`
	Version         string `json:"version,omitempty"`
	RequiresWP      string `json:"requires_wp,omitempty"`
	RequiresPHP     string `json:"requires_php,omitempty"`
	Author          string `json:"author,omitempty"`
	AuthorURI       string `json:"author_uri,omitempty"`
	License         string `json:"license,omitempty"`
	LicenseURI      string `json:"license_uri,omitempty"`
	UpdateURI       string `json:"update_uri,omitempty"`
	TextDomain      string `json:"text_domain,omitempty"`
	DomainPath      string `json:"domain_path,omitempty"`
	RequiresPlugins string `json:"requires_plugins,omitempty"`
	Network         string `json:"network,omitempty"`
}

// ParsePlugin reads from r and attempts to extract WordPress plugin headers. If
// a plugin name is found it returns a populated [Plugin] struct. Otherwise, it
// returns an error.
//
// The function mirrors WordPress [get_plugin_data] function:
//   - CR is normalized to LF
//   - best-effort encoding conversion is applied
//   - only the first 8 KiB is read
//
// [get_plugin_data]: https://developer.wordpress.org/reference/functions/get_plugin_data/
func ParsePlugin(r io.Reader) (Plugin, error) {
	s, err := read(r)
	if err != nil {
		return Plugin{}, err
	}

	name := extractHeader(s, "plugin_name")
	if name == "" {
		return Plugin{}, errNoHeader
	}

	return Plugin{
		Name:            name,
		URI:             extractHeader(s, "plugin_uri"),
		Description:     extractHeader(s, "description"),
		Version:         extractHeader(s, "version"),
		RequiresWP:      extractHeader(s, "requires_at_least"),
		RequiresPHP:     extractHeader(s, "requires_php"),
		Author:          extractHeader(s, "author"),
		AuthorURI:       extractHeader(s, "author_uri"),
		License:         extractHeader(s, "license"),
		LicenseURI:      extractHeader(s, "license_uri"),
		UpdateURI:       extractHeader(s, "update_uri"),
		TextDomain:      extractHeader(s, "text_domain"),
		DomainPath:      extractHeader(s, "domain_path"),
		RequiresPlugins: extractHeader(s, "requires_plugins"),
		Network:         extractHeader(s, "network"),
	}, nil
}

func read(r io.Reader) (string, error) {
	// Read first 8 KiB
	b, err := io.ReadAll(io.LimitReader(r, 8192))
	if err != nil {
		return "", fmt.Errorf("reading headers: %v", err)
	}

	bUTF8 := convertToUTF8(b)

	// Normalize CR to LF like WordPress
	s := strings.ReplaceAll(string(bUTF8), "\r", "\n")

	return s, nil
}

const (
	patternHead = `(?mi)^(?:[ \t]*<\?php)?[ \t\/*#@]*`
	patternTail = `:(.*)$`
)

var regexps = map[string]*regexp.Regexp{ //nolint:gochecknoglobals
	"plugin_name":       regexp.MustCompile(patternHead + regexp.QuoteMeta("Plugin Name") + patternTail),
	"plugin_uri":        regexp.MustCompile(patternHead + regexp.QuoteMeta("Plugin URI") + patternTail),
	"description":       regexp.MustCompile(patternHead + regexp.QuoteMeta("Description") + patternTail),
	"version":           regexp.MustCompile(patternHead + regexp.QuoteMeta("Version") + patternTail),
	"requires_at_least": regexp.MustCompile(patternHead + regexp.QuoteMeta("Requires at least") + patternTail),
	"requires_php":      regexp.MustCompile(patternHead + regexp.QuoteMeta("Requires PHP") + patternTail),
	"author":            regexp.MustCompile(patternHead + regexp.QuoteMeta("Author") + patternTail),
	"author_uri":        regexp.MustCompile(patternHead + regexp.QuoteMeta("Author URI") + patternTail),
	"license":           regexp.MustCompile(patternHead + regexp.QuoteMeta("License") + patternTail),
	"license_uri":       regexp.MustCompile(patternHead + regexp.QuoteMeta("License URI") + patternTail),
	"update_uri":        regexp.MustCompile(patternHead + regexp.QuoteMeta("Update URI") + patternTail),
	"text_domain":       regexp.MustCompile(patternHead + regexp.QuoteMeta("Text Domain") + patternTail),
	"domain_path":       regexp.MustCompile(patternHead + regexp.QuoteMeta("Domain Path") + patternTail),
	"requires_plugins":  regexp.MustCompile(patternHead + regexp.QuoteMeta("Requires Plugins") + patternTail),
	"network":           regexp.MustCompile(patternHead + regexp.QuoteMeta("Network") + patternTail),
	// Theme specific
	"theme_name":   regexp.MustCompile(patternHead + regexp.QuoteMeta("Theme Name") + patternTail),
	"theme_uri":    regexp.MustCompile(patternHead + regexp.QuoteMeta("Theme URI") + patternTail),
	"tags":         regexp.MustCompile(patternHead + regexp.QuoteMeta("Tags") + patternTail),
	"tested_up_to": regexp.MustCompile(patternHead + regexp.QuoteMeta("Tested up to") + patternTail),
	"template":     regexp.MustCompile(patternHead + regexp.QuoteMeta("Template") + patternTail),
}

// extractHeader searches s for a header using a regular expression and returns
// the trimmed captured value.
//
// It mirrors WordPress [get_file_data] function.
//
// Only default headers are supported.
//
// [get_file_data]: https://developer.wordpress.org/reference/functions/get_file_data/
func extractHeader(s, header string) string {
	re, ok := regexps[strings.ToLower(header)]
	if !ok {
		return ""
	}

	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return cleanupHeaderComment(m[1])
}

// cleanupHeaderComment removes comment terminators (*/ or ?>) and all following
// content on the same line, then trims surrounding whitespace.
//
// It mirrors WordPress [_cleanup_header_comment] function.
//
// [_cleanup_header_comment]: https://developer.wordpress.org/reference/functions/_cleanup_header_comment/
func cleanupHeaderComment(s string) string {
	s = strings.TrimSpace(s)

	// Remove comment terminators (*/ or ?>) and all following content,
	// mirroring WordPress _cleanup_header_comment semantics.
	// Two passes: first */ then ?>. If */ precedes ?>, the second pass is a
	// no-op; if ?> precedes */, the first pass truncates at */ and the second
	// pass then removes the exposed ?>.
	for _, term := range []string{"*/", "?>"} {
		if before, _, found := strings.Cut(s, term); found {
			s = strings.TrimRight(before, " \t")
		}
	}

	return strings.TrimSpace(s)
}
