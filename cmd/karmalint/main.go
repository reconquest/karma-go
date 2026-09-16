package main

import (
	"fmt"
	"os"

	"github.com/kovetskiy/docopt-go"
)

const usage = `karmalint - detect & fix common return error clauses.

Usage:
    karmalint [--fix] [<path>...]
    karmalint -h | --help

Options:
    --fix       Automatically fix issues where possible.
    -h --help   Show this help.

A path can be a Go file, a directory, or a recursive pattern like ./...
When no path is given, the current directory is checked.
`

func main() {
	opts, err := docopt.ParseArgs(usage, nil, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "karmalint:", err)
		os.Exit(1)
	}

	fix, err := opts.Bool("--fix")
	if err != nil {
		fmt.Fprintln(os.Stderr, "karmalint:", err)
		os.Exit(1)
	}

	paths, _ := opts["<path>"].([]string)

	files, err := collectFiles(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "karmalint:", err)
	}

	code := 0
	if fix {
		code = runWithFix(files)
	} else {
		code = runLintOnly(files)
	}

	if err != nil && code == 0 {
		code = 1
	}

	os.Exit(code)
}

func runLintOnly(files []string) int {
	issueCount := 0
	failed := false

	for _, file := range files {
		issues, err := lintFile(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "karmalint:", err)
			failed = true
			continue
		}

		for _, issue := range issues {
			printIssue(issue)
		}

		issueCount += len(issues)
	}

	if issueCount > 0 {
		fmt.Printf("Found %d issue(s)\n", issueCount)
	}

	if issueCount > 0 || failed {
		return 1
	}

	return 0
}

func runWithFix(files []string) int {
	fixedCount := 0
	failed := false

	for _, file := range files {
		fixed, err := fixFileInPlace(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "karmalint:", err)
			failed = true
			continue
		}

		if fixed > 0 {
			fmt.Printf("Fixed %d issue(s) in %s\n", fixed, file)
		}

		fixedCount += fixed
	}

	if fixedCount > 0 {
		fmt.Printf("Fixed %d issue(s)\n", fixedCount)
	}

	if failed {
		return 1
	}

	return 0
}

func printIssue(issue Issue) {
	fmt.Printf("%s:%d:%d: %s in %s()\n",
		issue.Filename,
		issue.Line,
		issue.Column,
		issue.Message,
		issue.FuncName,
	)
	fmt.Printf("    original:   %s\n", issue.Original)
	fmt.Printf("    suggestion: %s\n", issue.Suggestion)
	fmt.Println()
}
