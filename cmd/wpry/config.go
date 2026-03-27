package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const art = `
▗▖ ▗▖▗▄▄▖  ▄▄▄ ▄   ▄
▐▌ ▐▌▐▌ ▐▌█    █   █
▐▌ ▐▌▐▛▀▘ █     ▀▀▀█
▐▙█▟▌▐▌        ▄   █
                ▀▀▀
`

const ads = `
SUPPORT WPRY:
  If you find this tool useful, please consider supporting its development.
  Every contribution counts, regardless how big or small.
  I am eternally grateful to all sponsors who fund my open source journey.

GitHub Sponsor  https://github.com/sponsors/tangrufus

HIRE TANG RUFUS:
  I am looking for my next role, freelance or full-time.
  If you find this tool useful, I can build you more weird stuff like this.
  Let's talk if you are hiring PHP / Ruby / Go developers.

Contact         https://typist.tech/contact/
`

type config struct {
	parallel int
	timeout  time.Duration
}

func mustParseInput(args []string, stderr io.Writer) (string, config) {
	var cfg config

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	flags.Usage = func() {
		w := flags.Output()

		fmt.Fprintln(w, "USAGE:")
		fmt.Fprintf(w, "  %s [<flags>...] <path>\n", args[0])

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "FLAGS:")
		flags.PrintDefaults()

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "EXAMPLES:")
		fmt.Fprintln(w, "  # Parse a plugin main file")
		fmt.Fprintf(w, "  $ %s /path/to/index.php\n", args[0])
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  # Parse a theme main stylesheet")
		fmt.Fprintf(w, "  $ %s /path/to/style.css\n", args[0])
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  # Parse an unzipped plugin")
		fmt.Fprintf(w, "  $ %s /path/to/wp-content/plugins/woocommerce\n", args[0])
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  # Parse an unzipped theme")
		fmt.Fprintf(w, "  $ %s /path/to/wp-content/themes/twentytwentynine\n", args[0])

		fmt.Fprint(w, ads)
	}

	flags.IntVar(
		&cfg.parallel,
		"parallel",
		0,
		"run `n` workers simultaneously."+`
If n is 0 or less, GOMAXPROCS is used. Setting -parallel to values higher
 than GOMAXPROCS may cause degraded performance due to CPU contention.
(default GOMAXPROCS)`)

	flags.DurationVar(
		&cfg.timeout,
		"timeout",
		time.Minute,
		"If the parser runs longer than duration `d`, abort.",
	)

	var ver bool
	flags.BoolVar(&ver, "version", false, "Print version")
	flags.BoolVar(&ver, "v", false, "Print version")

	// Ignore error because of flag.ExitOnError
	err := flags.Parse(args[1:])
	if errors.Is(err, flag.ErrHelp) {
		flags.Usage()
		os.Exit(0)
	}

	if ver {
		printVersion(flags.Output())
		os.Exit(0)
	}

	if cfg.parallel < 1 {
		cfg.parallel = runtime.GOMAXPROCS(0)
	}

	positional := flags.Args()
	if len(positional) != 1 {
		fmt.Fprintf(stderr, "invalid positional arguments count: got %d, want %d\n", len(positional), 1)
		flags.Usage()
		os.Exit(2)
	}

	path := positional[0]
	if path == "" {
		fmt.Fprintf(stderr, "invalid value %q for path\n", path)
		flags.Usage()
		os.Exit(2)
	}

	return path, cfg
}

func printVersion(w io.Writer) {
	version := "(devel)"
	dirty := true
	var revision string

	if bi, ok := debug.ReadBuildInfo(); ok {
		version = bi.Main.Version

		for _, kv := range bi.Settings {
			switch kv.Key {
			case "vcs.modified":
				dirty = kv.Value != "false"
			case "vcs.revision":
				revision = kv.Value
			}
		}
	}

	url := "https://github.com/typisttech/wpry"
	switch {
	case strings.HasPrefix(version, "v") && strings.Count(version, "-") < 2:
		url = fmt.Sprintf("%s/releases/tag/%s", url, version)
	case !dirty && revision != "":
		url = fmt.Sprintf("%s/tree/%s", url, revision)
	}

	fmt.Fprint(w, art)

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "%-16s%s\n", "WPry", version)

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Parses WordPress plugin and theme headers.")
	fmt.Fprintln(w, url)

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Built with %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	fmt.Fprintln(w, "")
	fmt.Fprint(w, ads)
}
