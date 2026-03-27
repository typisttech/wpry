// Package wpry parses WordPress [plugin] and [theme] headers from files or
// filesystems.
//
// The parsers mirror WordPress behavior:
//   - [get_plugin_data]
//   - [wp_get_theme]
//   - [get_file_data]
//   - [_cleanup_header_comment]
//
// [plugin]: https://developer.wordpress.org/plugins/plugin-basics/header-requirements/
// [theme]: https://developer.wordpress.org/themes/classic-themes/basics/main-stylesheet-style-css/
// [get_plugin_data]: https://developer.wordpress.org/reference/functions/get_plugin_data/
// [wp_get_theme]: https://developer.wordpress.org/reference/functions/wp_get_theme/
// [get_file_data]: https://developer.wordpress.org/reference/functions/get_file_data/
// [_cleanup_header_comment]: https://developer.wordpress.org/reference/functions/_cleanup_header_comment/
package wpry
