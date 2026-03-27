package wpry

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"testing/synctest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseThemeFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fsys     fs.FS
		want     Theme
		wantPath string
		wantErr  bool
	}{
		{
			name: "finds-main-style-css",
			fsys: fstest.MapFS{
				"not-style.css": &fstest.MapFile{Data: []byte("/* Theme Name: Foo */")},
				"style.css":     &fstest.MapFile{Data: []byte("/* Theme Name: Bar */")},
				"readme.txt":    &fstest.MapFile{Data: []byte("/* Theme Name: Baz */")},
			},
			want:     Theme{Name: "Bar"},
			wantPath: "style.css",
		},
		{
			name: "no-header",
			fsys: fstest.MapFS{
				"style.css": &fstest.MapFile{Data: []byte("body { }")},
			},
			wantErr: true,
		},
		{
			name:    "empty-file",
			fsys:    fstest.MapFS{"style.css": &fstest.MapFile{Data: []byte{}}},
			wantErr: true,
		},
		{
			name: "no-style-css",
			fsys: fstest.MapFS{
				"not-style.css": &fstest.MapFile{Data: []byte("body { }")},
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
				"style.css": &fstest.MapFile{Data: []byte("/* Theme Name: My Theme */")},
			},
			want:     Theme{Name: "My Theme"},
			wantPath: "style.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotPath, err := ParseThemeFS(t.Context(), tt.fsys)

			if tt.wantErr && err == nil {
				t.Fatal("TestParseThemeFS() unexpected success")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("TestParseThemeFS() unexpected error: %v", err)
			}

			if gotPath != tt.wantPath {
				t.Errorf("TestParseThemeFS() path = %q, want %q", gotPath, tt.wantPath)
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("TestParseThemeFS() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseThemeFS_Cancellation(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		fsys := &blockingFS{
			MapFS: fstest.MapFS{
				"not-style.css": &fstest.MapFile{Data: []byte("/* Theme Name: Not My Theme */")},
				"style.css":     &fstest.MapFile{Data: []byte("/* Theme Name: My Theme */")},
			},
			blocks: map[string]chan struct{}{},
		}
		t.Cleanup(func() {
			fsys.unblockAll()
		})

		fsys.block("style.css")

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		type result struct {
			theme Theme
			path  string
			err   error
		}
		out := make(chan result, 1)

		go func() {
			defer close(out)

			theme, path, err := ParseThemeFS(ctx, fsys)
			out <- result{theme: theme, path: path, err: err}
		}()

		// Ensure the ParseThemeFS goroutine is being blocked.
		synctest.Wait()

		cancel()
		got := <-out

		if !errors.Is(got.err, context.Canceled) {
			t.Errorf("ParseThemeFS() error = %v, want %v", got.err, context.Canceled)
		}
	})
}
