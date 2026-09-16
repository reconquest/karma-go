package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixFileInPlace(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantFixedCount int
		wantFile       string
	}{
		{
			name: "fix multiple issues in place",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	if err != nil {
		return karma.Format(err, "failed to connect")
	}
	return karma.Format(err, "unable to read")
}`,
			wantFixedCount: 2,
			wantFile: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	if err != nil {
		return karma.Format(err, "connect")
	}
	return karma.Format(err, "read")
}`,
		},
		{
			name: "no issues to fix",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connect to database")
}`,
			wantFixedCount: 0,
			wantFile: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connect to database")
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(tmpFile, []byte(tt.input), 0644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			fixedCount, err := fixFileInPlace(tmpFile)
			if err != nil {
				t.Fatalf("fixFileInPlace: %v", err)
			}

			if fixedCount != tt.wantFixedCount {
				t.Errorf("fixed %d issues, want %d", fixedCount, tt.wantFixedCount)
			}

			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("read fixed file: %v", err)
			}

			got := strings.TrimSpace(string(content))
			want := strings.TrimSpace(tt.wantFile)
			if got != want {
				t.Errorf("file content mismatch:\ngot:\n%s\n\nwant:\n%s", got, want)
			}
		})
	}
}

func TestCollectFixes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantFixCount int
	}{
		{
			name: "single issue",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "failed to connect")
}`,
			wantFixCount: 1,
		},
		{
			name: "multiple issues",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	if err != nil {
		return karma.Format(err, "failed to a")
	}
	return karma.Format(err, "unable to b")
}`,
			wantFixCount: 2,
		},
		{
			name: "no issues",
			input: `package test

import "github.com/reconquest/karma-go"

func f(err error) error {
	return karma.Format(err, "connect")
}`,
			wantFixCount: 0,
		},
		{
			name: "fmt.Errorf wrap collected as fix",
			input: `package test

import "fmt"

func f(err error) error {
	return fmt.Errorf("failed to connect: %w", err)
}`,
			wantFixCount: 1,
		},
		{
			name: "nested calls do not produce overlapping fixes",
			input: `package test

import (
	"fmt"

	"github.com/reconquest/karma-go"
)

func f(err error) error {
	return fmt.Errorf("failed to wrap %s: %w", karma.Format(err, "failed to inner"), err)
}`,
			wantFixCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(tmpFile, []byte(tt.input), 0644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("read file: %v", err)
			}

			fixes, err := collectFixes(tmpFile, content)
			if err != nil {
				t.Fatalf("collectFixes: %v", err)
			}

			if len(fixes) != tt.wantFixCount {
				t.Errorf("got %d fixes, want %d", len(fixes), tt.wantFixCount)
			}

			for i, fix := range fixes {
				for j := i + 1; j < len(fixes); j++ {
					other := fixes[j]
					if fix.Start < other.End && other.Start < fix.End {
						t.Errorf("fixes %d and %d overlap", i, j)
					}
				}
			}
		})
	}
}
