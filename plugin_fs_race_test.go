//go:build !race

package wpry

import (
	"testing"
	"testing/fstest"
	"testing/synctest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
