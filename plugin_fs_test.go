package wpry

import (
	"context"
	"errors"
	"io/fs"
	"maps"
	"slices"
	"testing"
	"testing/fstest"
	"testing/synctest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type errReadDirFS struct{ err error }

func (e errReadDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, e.err
}

func (e errReadDirFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

var (
	_ fs.FS        = errReadDirFS{}
	_ fs.ReadDirFS = errReadDirFS{}
)

func TestParsePluginFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fsys     fs.FS
		opts     []ParseOption
		want     Plugin
		wantPath string
		wantErr  bool
	}{
		{
			name: "finds-main-plugin-file",
			fsys: fstest.MapFS{
				"not-plugin.php": &fstest.MapFile{Data: []byte("<?php // no header")},
				"index.php":      &fstest.MapFile{Data: []byte("<?php // Plugin Name: Foo")},
				"readme.txt":     &fstest.MapFile{Data: []byte("<?php // Plugin Name: Bar")},
			},
			want:     Plugin{Name: "Foo"},
			wantPath: "index.php",
		},
		{
			name: "no-header",
			fsys: fstest.MapFS{
				"index.php": &fstest.MapFile{Data: []byte("<?php // no header")},
			},
			wantErr: true,
		},
		{
			name: "empty-file",
			fsys: fstest.MapFS{
				"index.php": &fstest.MapFile{Data: []byte{}},
			},
			wantErr: true,
		},
		{
			name: "no-php-files",
			fsys: fstest.MapFS{
				"readme.txt": &fstest.MapFile{Data: []byte("<?php // Plugin Name: Foo")},
			},
			wantErr: true,
		},
		{
			name:    "no-files",
			fsys:    fstest.MapFS{},
			wantErr: true,
		},
		{
			name:    "read-dir-error",
			fsys:    errReadDirFS{err: errors.New("boom")},
			wantErr: true,
		},
		{
			name: "with-max-workers-option",
			fsys: fstest.MapFS{
				"index.php": &fstest.MapFile{Data: []byte("<?php // Plugin Name: My Plugin")},
			},
			opts:     []ParseOption{WithMaxWorkers(1)},
			want:     Plugin{Name: "My Plugin"},
			wantPath: "index.php",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotPath, err := ParsePluginFS(t.Context(), tt.fsys, tt.opts...)

			if tt.wantErr && err == nil {
				t.Fatal("ParsePluginFS() unexpected success")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("ParsePluginFS() unexpected error: %v", err)
			}

			if gotPath != tt.wantPath {
				t.Errorf("ParsePluginFS() path = %q, want %q", gotPath, tt.wantPath)
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ParsePluginFS() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParsePluginFS_MultipleMatch(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"not-plugin.php": &fstest.MapFile{Data: []byte("<?php // no header")},
		"foo.php":        &fstest.MapFile{Data: []byte("<?php // Plugin Name: Foo")},
		"bar.php":        &fstest.MapFile{Data: []byte("<?php // Plugin Name: Bar")},
		"baz.php":        &fstest.MapFile{Data: []byte("<?php // Plugin Name: Baz")},
		"qux.php":        &fstest.MapFile{Data: []byte("<?php // Plugin Name: Qux")},
	}

	got, gotPath, err := ParsePluginFS(t.Context(), fsys, WithMaxWorkers(1))
	if err != nil {
		t.Fatalf("ParsePluginFS() unexpected error: %v", err)
	}

	wants := map[string]Plugin{
		"foo.php": {Name: "Foo"},
		"bar.php": {Name: "Bar"},
		"baz.php": {Name: "Baz"},
		"qux.php": {Name: "Qux"},
	}

	want, ok := wants[gotPath]
	if !ok {
		t.Fatalf("ParsePluginFS() unexpected path: %q, want one of %q", gotPath, slices.Sorted(maps.Keys(wants)))
	}

	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("ParsePluginFS() mismatch (-want +got):\n%s", diff)
	}
}

type blockingFS struct {
	MapFS fstest.MapFS

	blocks map[string]chan struct{}
}

func (b blockingFS) Open(name string) (fs.File, error) {
	ch, ok := b.blocks[name]
	if ok {
		<-ch
	}

	return b.MapFS.Open(name)
}

func (b blockingFS) block(name string) {
	_, ok := b.blocks[name]
	if ok {
		return
	}

	b.blocks[name] = make(chan struct{})
}

func (b blockingFS) unblockAll() {
	for name, ch := range b.blocks {
		close(ch)
		delete(b.blocks, name)
	}
}

var _ fs.FS = blockingFS{}

func TestParsePluginFS_FastestWins(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		fsys := &blockingFS{
			MapFS: fstest.MapFS{
				"foo.php": &fstest.MapFile{Data: []byte("<?php // Plugin Name: Foo")},
				"bar.php": &fstest.MapFile{Data: []byte("<?php // Plugin Name: Bar")},
				"baz.php": &fstest.MapFile{Data: []byte("<?php // Plugin Name: Baz")},
			},
			blocks: map[string]chan struct{}{},
		}
		t.Cleanup(func() {
			fsys.unblockAll()
		})

		fsys.block("foo.php")
		fsys.block("baz.php")

		type result struct {
			plugin Plugin
			path   string
			err    error
		}

		out := make(chan result, 1)
		go func() {
			defer close(out)

			plugin, path, err := ParsePluginFS(t.Context(), fsys)
			out <- result{plugin: plugin, path: path, err: err}
		}()

		got := <-out

		if got.err != nil {
			t.Fatalf("ParsePluginFS() unexpected error: %v", got.err)
		}

		if got.path != "bar.php" {
			t.Errorf("ParsePluginFS() path = %q, want %q", got.path, "bar.php")
		}

		want := Plugin{Name: "Bar"}
		if diff := cmp.Diff(want, got.plugin, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("ParsePluginFS() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestParsePluginFS_Cancellation(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		fsys := &blockingFS{
			MapFS: fstest.MapFS{
				"not-plugin.php": &fstest.MapFile{Data: []byte("<?php // no header")},
				"foo.php":        &fstest.MapFile{Data: []byte("<?php // Plugin Name: Foo")},
			},
			blocks: map[string]chan struct{}{},
		}
		t.Cleanup(func() {
			fsys.unblockAll()
		})

		fsys.block("foo.php")

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		type result struct {
			plugin Plugin
			path   string
			err    error
		}

		out := make(chan result, 1)
		go func() {
			defer close(out)

			plugin, path, err := ParsePluginFS(ctx, fsys)
			out <- result{plugin: plugin, path: path, err: err}
		}()

		// Ensure the ParsePluginFS goroutine is being blocked.
		synctest.Wait()

		cancel()
		got := <-out

		if !errors.Is(got.err, context.Canceled) {
			t.Errorf("ParsePluginFS() error = %v, want %v", got.err, context.Canceled)
		}
	})
}
