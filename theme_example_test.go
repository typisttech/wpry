package wpry_test

import (
	"fmt"
	"strings"

	"github.com/typisttech/wpry"
)

func ExampleParseTheme() {
	css := `/*
Theme Name: Example Theme
Version: 1.2.3
*/`

	t, err := wpry.ParseTheme(strings.NewReader(css))
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(t.Name)
	fmt.Println(t.Version)
	// Output:
	// Example Theme
	// 1.2.3
}
