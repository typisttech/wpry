package wpry_test

import (
	"fmt"
	"strings"

	"github.com/typisttech/wpry"
)

func ExampleParsePlugin() {
	content := `<?php
/*
 * Plugin Name: Example Plugin
 * Version: 1.2.3
 */
`

	p, err := wpry.ParsePlugin(strings.NewReader(content))
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(p.Name)
	fmt.Println(p.Version)
	// Output:
	// Example Plugin
	// 1.2.3
}
