package main

import (
	"archive/zip"
	"io/fs"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

func openZipFS(path string) (fs.FS, func() error, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() error { return zr.Close() }

	var root string
	for _, f := range zr.File {
		p := f.Name
		if strings.HasSuffix(p, "/") {
			continue
		}

		p = strings.TrimLeft(p, "/")
		dir, _, found := strings.Cut(p, "/")
		if !found {
			root = ""
			break
		}

		if root != "" && dir != root {
			root = ""
			break
		}

		root = dir
	}

	var fsys fs.FS

	afs := zipfs.New(&zr.Reader)
	fsys = afero.NewIOFS(afs)
	if root != "" {
		fsys, err = fs.Sub(fsys, root)
		if err != nil {
			_ = cleanup()
			return nil, nil, err
		}
	}

	return fsys, cleanup, nil
}
