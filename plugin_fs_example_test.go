package wpry_test

import (
	"context"
	"fmt"
	"testing/fstest"

	"github.com/typisttech/wpry"
)

func ExampleParsePluginFS() {
	content := `<?php
/*
 * Plugin Name: Example Plugin
 * Version: 1.2.3
 */
`

	fsys := fstest.MapFS{
		"plugin.php": {Data: []byte(content)},
	}

	p, path, err := wpry.ParsePluginFS(context.Background(), fsys)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(path)
	fmt.Println(p.Name)
	fmt.Println(p.Version)
	// Output:
	// plugin.php
	// Example Plugin
	// 1.2.3
}
