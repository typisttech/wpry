package wpry

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

// ParseThemeFS scans the main Stylesheet (style.css) under fsys (style.css) and
// attempts to parse its headers. If successful, a [Theme] struct and its path
// are returned. Otherwise, it returns an error.
func ParseThemeFS(ctx context.Context, fsys fs.FS) (Theme, string, error) {
	type result struct {
		theme Theme
		path  string
		err   error
	}
	out := make(chan result, 1)

	go func() {
		defer close(out)

		ents, err := fs.ReadDir(fsys, ".")
		if err != nil {
			out <- result{err: fmt.Errorf("reading directory: %v", err)}
			return
		}

		var name string
		for _, ent := range ents {
			if ent.IsDir() {
				continue
			}
			if strings.EqualFold(ent.Name(), "style.css") {
				name = ent.Name()
				break
			}
		}

		if name == "" {
			out <- result{err: errors.New("style.css not found")}
			return
		}

		f, err := fsys.Open(name)
		if err != nil {
			out <- result{err: fmt.Errorf("opening %s: %v", name, err)}
			return
		}
		defer f.Close()

		t, err := ParseTheme(f)
		if err != nil {
			out <- result{err: err}
			return
		}

		out <- result{theme: t, path: name}
	}()

	select {
	case r := <-out:
		return r.theme, r.path, r.err
	case <-ctx.Done():
		var zero Theme
		return zero, "", ctx.Err()
	}
}
