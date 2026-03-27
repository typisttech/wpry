package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"testing"
)

func TestOpenZipFS(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		files  map[string]string
		target string
		want   string
	}{
		{
			name: "single-file",
			files: map[string]string{
				"foo.txt": "Foo",
			},
			target: "foo.txt",
			want:   "Foo",
		},
		{
			name: "strip-top-level",
			files: map[string]string{
				"foo/bar.txt": "Bar",
				"foo/baz.txt": "Baz",
			},
			target: "bar.txt",
			want:   "Bar",
		},
		{
			name: "top-level-file",
			files: map[string]string{
				"foo.txt":     "Foo",
				"bar/baz.txt": "Baz",
			},
			target: "foo.txt",
			want:   "Foo",
		},
		{
			name: "no-top-level-file",
			files: map[string]string{
				"foo/bar.txt": "Bar",
				"baz/quz.txt": "Quz",
			},
			target: "foo/bar.txt",
			want:   "Bar",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := createZip(t, tt.files)

			fsys, cleanup, err := openZipFS(path)
			if err != nil {
				t.Fatalf("openZipFS() unexpected error: %v", err)
			}
			t.Cleanup(func() { _ = cleanup() })

			b, err := fs.ReadFile(fsys, tt.target)
			if err != nil {
				t.Fatalf("fs.ReadFile(fsys, %q) unexpected error: %v", tt.target, err)
			}
			got := string(b)
			if got != tt.want {
				t.Fatalf("fs.ReadFile(fsys, %q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func createZip(t *testing.T, files map[string]string) string {
	t.Helper()

	tmp := t.TempDir()
	root, err := os.OpenRoot(tmp)
	if err != nil {
		t.Fatalf("os.OpenRoot(%q) unexpected error: %v", tmp, err)
	}

	f, err := root.Create("test.zip")
	if err != nil {
		t.Fatalf("os.Create(%q) unexpected error: %v", "test.zip", err)
	}

	zw := zip.NewWriter(f)

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zw.Create(%q) unexpected error: %v", name, err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatalf("io.WriteString() for %q unexpected error: %v", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zw.Close() unexpected error: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("f.Close() unexpected error: %v", err)
	}

	return f.Name()
}
