package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinter(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantIssueCount int
		wantMessages   []string
		wantFixed      string
	}{
		{
			name: "failed to prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "failed to connect to database")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"failed to" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connect to database")
}`,
		},
		{
			name: "unable to prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "unable to read file")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"unable to" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "read file")
}`,
		},
		{
			name: "could not prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "could not parse config")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"could not" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "parse config")
}`,
		},
		{
			name: "cannot prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "cannot open file")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"cannot" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "open file")
}`,
		},
		{
			name: "can't prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "can't find user")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"can't" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "find user")
}`,
		},
		{
			name: "error prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "error connecting to server")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"error" prefix is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connecting to server")
}`,
		},
		{
			name: "err: prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "err: something went wrong")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"err:" prefix is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "something went wrong")
}`,
		},
		{
			name: "Failed to uppercase",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "Failed to connect")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"failed to" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connect")
}`,
		},
		{
			name: "trailing failed",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "database connection failed")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"trailing failure word is redundant"},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "database connection")
}`,
		},
		{
			name: "trailing error",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "parse config error")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"trailing failure word is redundant"},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "parse config")
}`,
		},
		{
			name: "Describe chain with redundant prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(path string, err error) error {
	return karma.Describe("path", path).Format(err, "failed to open config")
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"failed to" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(path string, err error) error {
	return karma.Describe("path", path).Format(err, "open config")
}`,
		},
		{
			name: "Collect with redundant prefix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(errs ...error) error {
	return karma.Collect("failed to validate config", errs...)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{`"failed to" is redundant`},
			wantFixed: `package test

import "github.com/reconquest/karma-go"

func f(errs ...error) error {
	return karma.Collect("validate config", errs...)
}`,
		},
		{
			name: "valid karma.Format message",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connect to database")
}`,
			wantIssueCount: 0,
		},
		{
			name: "valid karma.Format with args",
			input: `package test

import "github.com/reconquest/karma-go"

func f(name string, err error) error {
	return karma.Format(err, "process %s", name)
}`,
			wantIssueCount: 0,
		},
		{
			name: "empty string message",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "")
}`,
			wantIssueCount: 0,
		},
		{
			name: "message in middle of text",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "the system failed to connect due to timeout")
}`,
			wantIssueCount: 0,
		},
		{
			name: "non-karma Format call is ignored",
			input: `package test
func f(f Formatter, err error) error {
	return f.Format(err, "failed to connect")
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with %w",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("failed to fetch volume data: %w", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format instead of fmt.Errorf"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return karma.Format(err, "fetch volume data")
}`,
		},
		{
			name: "fmt.Errorf with %w and trailing failed",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("database query failed: %w", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return karma.Format(err, "database query")
}`,
		},
		{
			name: "fmt.Errorf with %w and multiple args",
			input: `package test

import "fmt"

func f(name string, err error) error {
	return fmt.Errorf("failed to process %s: %w", name, err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(name string, err error) error {
	return karma.Format(err, "process %s", name)
}`,
		},
		{
			name: "fmt.Errorf with %w and complex arg",
			input: `package test

import "fmt"

func f(m map[string]int, err error) error {
	return fmt.Errorf("failed to process %d: %w", m["key"], err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(m map[string]int, err error) error {
	return karma.Format(err, "process %d", m["key"])
}`,
		},
		{
			name: "fmt.Errorf with selector error",
			input: `package test

import "fmt"

func f(state State) error {
	return fmt.Errorf("failed to open: %w", state.Err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(state State) error {
	return karma.Format(state.Err, "open")
}`,
		},
		{
			name: "fmt.Errorf with quoted message",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("failed to parse \"config\" section: %w", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return karma.Format(err, "parse \"config\" section")
}`,
		},
		{
			name: "fmt.Errorf with escaped percent",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("coverage 100%%: %w", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return karma.Format(err, "coverage 100%%")
}`,
		},
		{
			name: "fmt.Errorf already importing karma",
			input: `package test

import (
	"fmt"

	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return fmt.Errorf("failed to connect: %w", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return karma.Format(err, "connect")
}`,
		},
		{
			name: "fmt.Errorf with customErr variable",
			input: `package test

import "fmt"

func f(customErr error) error {
	return fmt.Errorf("operation failed: %w", customErr)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(customErr error) error {
	return karma.Format(customErr, "operation")
}`,
		},
		{
			name: "fmt.Errorf with %v loses error chain",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("failed to connect: %v", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"keep the error chain"},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return karma.Format(err, "connect")
}`,
		},
		{
			name: "fmt.Errorf without trailing error verb",
			input: `package test

import "fmt"

func f(name string) error {
	return fmt.Errorf("invalid name: %s", name)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with %w in middle",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("wrapped %w occurred during processing", err)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with two wrapping verbs",
			input: `package test

import "fmt"

func f(errOne, errTwo error) error {
	return fmt.Errorf("first %w: %w", errOne, errTwo)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with variadic args",
			input: `package test

import "fmt"

func f(errs ...error) error {
	return fmt.Errorf("open: %w", errs...)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with mismatched verb count",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("values %d %s: %w", err)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with non-error variable",
			input: `package test

import "fmt"

func f(msg string) error {
	return fmt.Errorf("result: %w", msg)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with errorText variable",
			input: `package test

import "fmt"

func f(errorText string) error {
	return fmt.Errorf("result: %v", errorText)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Sprintf is not flagged",
			input: `package test

import "fmt"

func f(err error) string {
	return fmt.Sprintf("failed to connect: %s", err)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt import kept when still used",
			input: `package test

import "fmt"

func f(err error) error {
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	fmt.Println("ok")
	return nil
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"fmt"

	"github.com/reconquest/karma-go"
)

func f(err error) error {
	if err != nil {
		return karma.Format(err, "connect")
	}
	fmt.Println("ok")
	return nil
}`,
		},
		{
			name: "unformatted file keeps its layout",
			input: `package test

import "fmt"

func f(err error)  error {
	return fmt.Errorf("failed to connect: %w", err)
}`,
			wantIssueCount: 1,
			wantMessages:   []string{"use karma.Format"},
			wantFixed: `package test

import (
	"fmt"
	"github.com/reconquest/karma-go"
)

func f(err error)  error {
	return karma.Format(err, "connect")
}`,
		},
		{
			name: "fmt.Errorf inside karma package is left alone",
			input: `package karma

import "fmt"

func f(err error) error {
	return fmt.Errorf("failed to connect: %w", err)
}`,
			wantIssueCount: 0,
		},
		{
			name: "fmt.Errorf with aliased karma import is left alone",
			input: `package test

import (
	"fmt"

	k "github.com/reconquest/karma-go"
)

func f(err error) error {
	return fmt.Errorf("failed to connect: %w", err)
}`,
			wantIssueCount: 0,
		},
		{
			name: "multiple issues",
			input: `package test

import (
	"fmt"

	"github.com/reconquest/karma-go"
)

func f(err error) error {
	if err != nil {
		return karma.Format(err, "failed to connect")
	}
	return fmt.Errorf("unable to start: %w", err)
}`,
			wantIssueCount: 2,
			wantMessages: []string{
				`"failed to" is redundant`,
				"use karma.Format",
			},
			wantFixed: `package test

import (
	"github.com/reconquest/karma-go"
)

func f(err error) error {
	if err != nil {
		return karma.Format(err, "connect")
	}
	return karma.Format(err, "start")
}`,
		},
		{
			name:           "backtick string",
			input:          "package test\nimport \"github.com/reconquest/karma-go\"\nfunc f(err error) error {\n\treturn karma.Format(err, `failed to connect`)\n}",
			wantIssueCount: 1,
			wantMessages:   []string{`"failed to" is redundant`},
			wantFixed:      "package test\nimport \"github.com/reconquest/karma-go\"\nfunc f(err error) error {\n\treturn karma.Format(err, `connect`)\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(tmpFile, []byte(tt.input), 0644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			issues, err := lintFile(tmpFile)
			if err != nil {
				t.Fatalf("lintFile: %v", err)
			}

			if len(issues) != tt.wantIssueCount {
				t.Errorf("got %d issues, want %d", len(issues), tt.wantIssueCount)
				for i, issue := range issues {
					t.Logf("  issue %d: %s", i+1, issue.Message)
				}
			}

			for _, wantMsg := range tt.wantMessages {
				found := false
				for _, issue := range issues {
					if strings.Contains(issue.Message, wantMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue message containing %q not found", wantMsg)
				}
			}

			if tt.wantFixed != "" {
				fixed, err := fixFile(tmpFile)
				if err != nil {
					t.Fatalf("fixFile: %v", err)
				}

				gotFixed := strings.TrimSpace(fixed)
				wantFixed := strings.TrimSpace(tt.wantFixed)

				if gotFixed != wantFixed {
					t.Errorf("fix mismatch:\ngot:\n%s\n\nwant:\n%s", gotFixed, wantFixed)
				}
			}
		})
	}
}

func TestSuggestFix(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		patternIndex int
		want         string
	}{
		{"failed to", "failed to connect", 0, "connect"},
		{"Failed to caps", "Failed to Connect", 0, "connect"},
		{"unable to", "unable to read", 1, "read"},
		{"could not", "could not parse", 2, "parse"},
		{"cannot", "cannot open", 3, "open"},
		{"can't", "can't find", 4, "find"},
		{"error prefix", "error connecting", 5, "connecting"},
		{"err: prefix", "err: something", 6, "something"},
		{"empty", "", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestFix(tt.input, redundantPatterns[tt.patternIndex].pattern)
			if got != tt.want {
				t.Errorf("suggestFix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanMessage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"failed to connect", "connect"},
		{"connect to database", "connect to database"},
		{"database query failed", "database query"},
		{"failed to connect failed", "connect"},
		{"parse config error", "parse config"},
		{"unable to read file err", "read file"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanMessage(tt.input)
			if got != tt.want {
				t.Errorf("cleanMessage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCollectFilesRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"root.go":           `package root`,
		"subdir/sub.go":     `package subdir`,
		".hidden/hidden.go": `package hidden`,
		"vendor/v.go":       `package vendor`,
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	found, err := collectFiles([]string{tmpDir + "/..."})
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("got %d files, want 2: %v", len(found), found)
	}

	for _, file := range found {
		if strings.Contains(file, ".hidden") || strings.Contains(file, "vendor") {
			t.Errorf("directory should be skipped: %s", file)
		}
	}
}
