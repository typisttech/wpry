package wpry

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
)

// ParseOption configures parse helpers.
type ParseOption func(*parseOptions)

type parseOptions struct {
	maxWorkers int
}

func WithMaxWorkers(n int) ParseOption {
	return func(o *parseOptions) {
		o.maxWorkers = n
	}
}

func ParsePluginFS(ctx context.Context, fsys fs.FS, opts ...ParseOption) (Plugin, string, error) { //nolint:cyclop
	var po parseOptions
	for _, opt := range opts {
		opt(&po)
	}
	if po.maxWorkers <= 0 {
		po.maxWorkers = runtime.GOMAXPROCS(0)
	}

	wgCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		plugin Plugin
		path   string
		err    error
	}
	out := make(chan result, 1)

	var wg sync.WaitGroup

	paths := make(chan string)

	wg.Go(func() {
		defer close(paths)

		ents, err := fs.ReadDir(fsys, ".")
		if err != nil {
			out <- result{err: err}
			return
		}
		names := slices.Collect(func(yield func(string) bool) {
			for _, ent := range ents {
				if ent.IsDir() {
					continue
				}

				name := ent.Name()
				ext := filepath.Ext(name)
				if !strings.EqualFold(ext, ".php") {
					continue
				}

				if !yield(name) {
					return
				}
			}
		})

		var found bool
		for _, name := range names {
			found = true

			select {
			case paths <- name:
			case <-wgCtx.Done():
				return
			}
		}

		if !found {
			out <- result{err: errors.New("PHP files not found")}
		}
	})

	for range po.maxWorkers {
		wg.Go(func() {
			for {
				select {
				case path, ok := <-paths:
					if !ok {
						return
					}

					f, err := fsys.Open(path)
					if err != nil {
						continue
					}
					defer f.Close()

					p, err := ParsePlugin(f)
					// Immediately close the file to free up resources to
					// prevent cumulating resources until Go routine exit.
					_ = f.Close()
					if err != nil {
						continue
					}

					select {
					case out <- result{plugin: p, path: path}:
						// Signal other WaitGroup routines to stop.
						cancel()
						return
					case <-wgCtx.Done():
						return
					}
				case <-wgCtx.Done():
					return
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	select {
	case re, ok := <-out:
		if !ok {
			var zero Plugin
			return zero, "", errors.New("main plugin PHP file not found")
		}
		if re.err != nil {
			var zero Plugin
			return zero, "", re.err
		}
		return re.plugin, re.path, nil
	case <-ctx.Done():
		var zero Plugin
		return zero, "", ctx.Err()
	}
}
