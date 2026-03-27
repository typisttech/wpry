package wpry_test

import (
	"context"
	"fmt"
	"testing/fstest"

	"github.com/typisttech/wpry"
)

func ExampleParseThemeFS() {
	content := `/*
Theme Name: Example Theme
Version: 1.2.3
*/`

	fsys := fstest.MapFS{
		"style.css": {Data: []byte(content)},
	}

	t, path, err := wpry.ParseThemeFS(context.Background(), fsys)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(path)
	fmt.Println(t.Name)
	fmt.Println(t.Version)
	// Output:
	// style.css
	// Example Theme
	// 1.2.3
}
