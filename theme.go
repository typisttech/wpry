package wpry

import (
	"io"
)

// Theme represents parsed theme metadata.
type Theme struct {
	Name        string `json:"name,omitempty"`
	URI         string `json:"uri,omitempty"`
	Author      string `json:"author,omitempty"`
	AuthorURI   string `json:"author_uri,omitempty"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	RequiresWP  string `json:"requires_wp,omitempty"`
	TestedUpTo  string `json:"tested_up_to,omitempty"`
	RequiresPHP string `json:"requires_php,omitempty"`
	License     string `json:"license,omitempty"`
	LicenseURI  string `json:"license_uri,omitempty"`
	TextDomain  string `json:"text_domain,omitempty"`
	DomainPath  string `json:"domain_path,omitempty"`
	Tags        string `json:"tags,omitempty"`
	Template    string `json:"template,omitempty"`
}

func ParseTheme(r io.Reader) (Theme, error) {
	s, err := read(r)
	if err != nil {
		return Theme{}, err
	}

	name := extractHeader(s, "theme_name")
	if name == "" {
		return Theme{}, errNoHeader
	}

	return Theme{
		Name:        name,
		URI:         extractHeader(s, "theme_uri"),
		Description: extractHeader(s, "description"),
		Version:     extractHeader(s, "version"),
		RequiresWP:  extractHeader(s, "requires_at_least"),
		RequiresPHP: extractHeader(s, "requires_php"),
		Author:      extractHeader(s, "author"),
		AuthorURI:   extractHeader(s, "author_uri"),
		License:     extractHeader(s, "license"),
		LicenseURI:  extractHeader(s, "license_uri"),
		TextDomain:  extractHeader(s, "text_domain"),
		DomainPath:  extractHeader(s, "domain_path"),
		TestedUpTo:  extractHeader(s, "tested_up_to"),
		Tags:        extractHeader(s, "tags"),
		Template:    extractHeader(s, "template"),
	}, nil
}
