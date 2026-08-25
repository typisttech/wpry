package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/typisttech/wpry"
)

var errNoHeader = errors.New("no header found")

func main() {
	ctx := context.Background()

	err := run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(1)
	}
}

type result struct {
	File   string      `json:"file,omitzero"`
	Plugin wpry.Plugin `json:"plugin,omitzero"`
	Theme  wpry.Theme  `json:"theme,omitzero"`
	Err    error       `json:"-"`
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	path, cfg := mustParseInput(args, stderr)

	ctx, cancel := context.WithTimeout(ctx, cfg.timeout)
	defer cancel()

	//gosec:disable G703 -- Obey users' intentions.
	fi, err := os.Stat(path)
	if err != nil {
		render(stdout, result{Err: err})
		return err
	}

	out := make(chan result, 1)

	go func() {
		var res result

		switch {
		case fi.IsDir():
			res = runDir(ctx, os.DirFS(path), cfg)
		case filepath.Ext(path) == ".zip":
			fsys, cleanup, err := openZipFS(path)
			if err != nil {
				res = result{Err: fmt.Errorf("invalid zip: %v", err)}
				break
			}
			defer func() { _ = cleanup() }()

			res = runDir(ctx, fsys, cfg)
		default:
			res = runFile(ctx, path)
		}

		select {
		case out <- res:
		case <-ctx.Done():
			out <- result{Err: ctx.Err()}
		}
	}()

	var re result
	select {
	case re = <-out:
	case <-ctx.Done():
		re = result{Err: ctx.Err()}
	}

	render(stdout, re)
	return re.Err
}

func runDir(ctx context.Context, fsys fs.FS, cfg config) result {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make(chan result, 2)

	var wg sync.WaitGroup

	wg.Go(func() {
		p, pp, err := wpry.ParsePluginFS(ctx, fsys, wpry.WithMaxWorkers(cfg.parallel))
		if err != nil {
			return
		}

		out <- result{Plugin: p, File: pp}
	})

	wg.Go(func() {
		t, tp, err := wpry.ParseThemeFS(ctx, fsys)
		if err != nil {
			return
		}

		out <- result{Theme: t, File: tp}
	})

	go func() {
		wg.Wait()
		close(out)
	}()

	select {
	case re, ok := <-out:
		if !ok {
			return result{Err: errNoHeader}
		}
		return re
	case <-ctx.Done():
		return result{Err: ctx.Err()}
	}
}

func runFile(ctx context.Context, path string) result {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	file := filepath.Base(path)

	out := make(chan result, 2)

	var wg sync.WaitGroup

	wg.Go(func() {
		//gosec:disable G304,G703 -- Obey users' intentions.
		r, err := os.Open(path)
		if err != nil {
			return
		}
		defer r.Close()

		p, err := wpry.ParsePlugin(r)
		if err != nil {
			return
		}

		out <- result{Plugin: p, File: file}
	})

	wg.Go(func() {
		//gosec:disable G304,G703 -- Obey users' intentions.
		r, err := os.Open(path)
		if err != nil {
			return
		}
		defer r.Close()

		t, err := wpry.ParseTheme(r)
		if err != nil {
			return
		}

		out <- result{Theme: t, File: file}
	})

	go func() {
		wg.Wait()
		close(out)
	}()

	select {
	case re, ok := <-out:
		if !ok {
			return result{Err: errNoHeader}
		}
		return re
	case <-ctx.Done():
		return result{Err: ctx.Err()}
	}
}

func render(out io.Writer, re result) {
	var v any = re
	if re.Err != nil {
		type erratum struct {
			Msg string `json:"error,omitzero"`
		}

		v = erratum{Msg: re.Err.Error()}
	}

	b, _ := json.Marshal(v) //nolint:errchkjson
	_, _ = out.Write(b)
	_, _ = out.Write([]byte("\n"))
}
