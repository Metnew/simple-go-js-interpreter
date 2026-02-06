package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/example/jsgo/testrunner"
)

func main() {
	test262Dir := flag.String("dir", "test262", "path to test262 checkout")
	filter := flag.String("filter", "", "filter tests by path substring")
	limit := flag.Int("limit", 0, "maximum number of tests to run (0 = all)")
	verbose := flag.Bool("v", false, "verbose output (print each test result)")
	flag.Parse()

	// Check test262 dir exists
	if _, err := os.Stat(*test262Dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: test262 directory not found at %s\n", *test262Dir)
		fmt.Fprintf(os.Stderr, "Clone it with: git clone --depth 1 https://github.com/nicolo-ribaudo/tc39-test262-parser %s\n", *test262Dir)
		os.Exit(1)
	}

	cfg := testrunner.Config{
		Test262Dir: *test262Dir,
		Filter:     *filter,
		Limit:      *limit,
		Verbose:    *verbose,
	}

	results, summary := testrunner.Run(cfg)

	// Print non-verbose results
	if !*verbose {
		for _, r := range results {
			msg := ""
			if r.Message != "" {
				msg = " " + r.Message
			}
			fmt.Printf("%s %s%s\n", r.Result, r.Path, msg)
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("=== Test262 Summary ===")
	fmt.Printf("Total:   %d\n", summary.Total)
	fmt.Printf("Passed:  %d\n", summary.Passed)
	fmt.Printf("Failed:  %d\n", summary.Failed)
	fmt.Printf("Skipped: %d\n", summary.Skipped)
	fmt.Printf("Errors:  %d\n", summary.Errors)
	if summary.Total > 0 {
		fmt.Printf("Pass rate: %.1f%% (%d/%d excluding skipped)\n",
			float64(summary.Passed)/float64(summary.Total-summary.Skipped)*100,
			summary.Passed,
			summary.Total-summary.Skipped)
	}
	fmt.Printf("Elapsed: %s\n", summary.Elapsed)

	if summary.Failed > 0 || summary.Errors > 0 {
		os.Exit(1)
	}
}
