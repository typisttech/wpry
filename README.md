<div align="center">

# WPry

[![Go Reference](https://pkg.go.dev/badge/github.com/typisttech/wpry.svg)](https://pkg.go.dev/github.com/typisttech/wpry)
[![GitHub Release](https://img.shields.io/github/v/release/typisttech/wpry?style=flat-square&)](https://github.com/typisttech/wpry/releases/latest)
[![Test](https://github.com/typisttech/wpry/actions/workflows/test.yml/badge.svg)](https://github.com/typisttech/wpry/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/typisttech/wpry/graph/badge.svg?token=HD0PRS6E1A)](https://codecov.io/gh/typisttech/wpry)
[![License](https://img.shields.io/github/license/typisttech/wpry.svg)](https://github.com/typisttech/wpry/blob/master/LICENSE)
[![Follow @TangRufus on X](https://img.shields.io/badge/Follow-TangRufus-15202B?logo=x&logoColor=white)](https://x.com/tangrufus)
[![Follow @TangRufus.com on Bluesky](https://img.shields.io/badge/Bluesky-TangRufus.com-blue?logo=bluesky)](https://bsky.app/profile/tangrufus.com)
[![Sponsor @TangRufus via GitHub](https://img.shields.io/badge/Sponsor-TangRufus-EA4AAA?logo=githubsponsors)](https://github.com/sponsors/tangrufus)
[![Hire Typist Tech](https://img.shields.io/badge/Hire-Typist%20Tech-778899)](https://typist.tech/contact/)

<p>
  <strong>WPry parses WordPress plugin and theme headers.</strong>
  <br />
  <br />
  Built with ♥ by <a href="https://typist.tech/">Typist Tech</a>
</p>

</div>

---

> [!TIP]
> **Hire Tang Rufus!**
>
> I am looking for my next role, freelance or full-time.
> If you find this tool useful, I can build you more weird stuff like this.
> Let's talk if you are hiring PHP / Ruby / Go developers.
>
> Contact me at https://typist.tech/contact/

---

## Features

- Parse WordPress plugin headers from PHP files
- Parse WordPress theme headers from `style.css`
- Parse unzipped plugin and theme directories
- Parse plugin and theme zip archives containing (CLI only)
- Normalize CR and CRLF line endings
- Apply best-effort encoding fallback before header parsing
  Heuristics (in order):
    1. If input is already valid UTF-8, return it (strip UTF-8 BOM if present)
    2. Check for UTF-32 BOMs (BE/LE) and decode when present
    3. Check for UTF-16 BOMs (BE/LE) and decode when present
    4. Try Windows-1252
    5. Try ISO-8859-1

## Library Usage

[![Go Reference](https://pkg.go.dev/badge/github.com/typisttech/wpry.svg)](https://pkg.go.dev/github.com/typisttech/wpry)

Refer to [Go Reference on pkg.go.dev](https://pkg.go.dev/github.com/typisttech/wpry#section-documentation)

## CLI Usage

```bash
USAGE:
  wpry [<flags>...] <path>

FLAGS:
  -parallel n
    	run n workers simultaneously.
    	If n is 0 or less, GOMAXPROCS is used. Setting -parallel to values higher
    	 than GOMAXPROCS may cause degraded performance due to CPU contention.
    	(default GOMAXPROCS)
  -timeout d
    	If the parser runs longer than duration d, abort. (default 1m0s)
  -v	Print version
  -version
    	Print version

EXAMPLES:
  # Parse a plugin main file
  $ wpry /path/to/index.php

  # Parse a theme main stylesheet
  $ wpry /path/to/style.css

  # Parse an unzipped plugin
  $ wpry /path/to/wp-content/plugins/woocommerce

  # Parse an unzipped theme
  $ wpry /path/to/wp-content/themes/twentytwentynine

  # Parse a plugin zip
  $ wpry /path/to/woocommerce.zip

  # Parse a theme zip
  $ wpry /path/to/twentytwentynine.zip
```

> [!TIP]
> **Hire Tang Rufus!**
>
> There is no need to understand any of these quirks.
> Let me handle them for you.
> I am seeking my next job, freelance or full-time.
>
> If you are hiring PHP / Ruby / Go developers,
> contact me at https://typist.tech/contact/

### Examples

Parse a plugin file:

```bash
wpry /path/to/index.php
```

```json
{
  "file": "index.php",
  "plugin": {
    "name": "Full Plugin",
    "uri": "https://example.com/full-plugin",
    "description": "A fully specified plugin.",
    "version": "2.0.0",
    "requires_wp": "6.0",
    "requires_php": "8.0",
    "author": "Full Author",
    "author_uri": "https://example.com/author",
    "license": "GPL-2.0-or-later",
    "license_uri": "https://www.gnu.org/licenses/gpl-2.0.html",
    "update_uri": "https://example.com/update",
    "text_domain": "full-plugin",
    "domain_path": "/lang",
    "requires_plugins": "woocommerce, akismet",
    "network": "true"
  }
}
```

Parse a theme main stylesheet:

```bash
wpry /path/to/style.css
```

```json
{
  "file": "style.css",
  "theme": {
    "name": "Full Theme",
    "uri": "https://example.com/full-theme",
    "author": "Full Author",
    "author_uri": "https://example.com/author",
    "description": "A fully specified theme.",
    "version": "3.0.0",
    "requires_wp": "6.0",
    "tested_up_to": "6.5",
    "requires_php": "8.0",
    "license": "GPL-2.0-or-later",
    "license_uri": "https://www.gnu.org/licenses/gpl-2.0.html",
    "text_domain": "full-theme",
    "domain_path": "/lang",
    "tags": "custom-background, custom-logo",
    "template": "twentytwentyfive"
  }
}
```

Parse a directory:

```bash
wpry /path/to/wp-content/plugins/woocommerce
wpry /path/to/wp-content/themes/twentytwentynine
```

Parse a zip archive:

```bash
wpry /path/to/woocommerce.zip
wpry /path/to/twentytwentynine.zip
```

Error result:

```json
{
  "error": "no header found"
}
```

### CLI Installation

#### Homebrew (macOS / Linux) (Recommended)

```bash
brew install typisttech/tap/wpry
```

#### Build from Source

```bash
go install github.com/typisttech/wpry/cmd/wpry@latest
```

#### Linux (Debian & Alpine)

Follow the instructions on https://broadcasts.cloudsmith.com/typisttech/oss

![Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square&link=https%3A%2F%2Fcloudsmith.com)

Package repository hosting is graciously provided by [Cloudsmith](https://cloudsmith.com).
Cloudsmith is the only fully hosted, cloud-native, universal package management solution, that
enables your organization to create, store and share packages in any format, to any place, with total
confidence.

## Credits

[WPry](https://github.com/typisttech/wpry) is a [Typist Tech](https://typist.tech) project and maintained by [Tang Rufus](https://x.com/TangRufus), freelance developer for [hire](https://typist.tech/contact/).

Full list of contributors can be found [here](https://github.com/typisttech/wpry/graphs/contributors).

## Copyright and License

This project is a [free software](https://www.gnu.org/philosophy/free-sw.en.html) distributed under the terms of the MIT license. For the full license, see [LICENSE](./LICENSE).

## Contribute

Feedbacks / bug reports / pull requests are welcome.
