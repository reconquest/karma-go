package main

import "testing"

func TestEnsureKarmaImport(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "no karma reference",
			input: `package test
import "fmt"
func f() { fmt.Println("x") }
`,
			want: `package test
import "fmt"
func f() { fmt.Println("x") }
`,
		},
		{
			name: "already imported",
			input: `package test
import "github.com/reconquest/karma-go"
func f(err error) error { return karma.Format(err, "x") }
`,
			want: `package test
import "github.com/reconquest/karma-go"
func f(err error) error { return karma.Format(err, "x") }
`,
		},
		{
			name: "aliased import left untouched",
			input: `package test
import k "github.com/reconquest/karma-go"
func f(err error) error { return k.Format(err, "x") }
`,
			want: `package test
import k "github.com/reconquest/karma-go"
func f(err error) error { return k.Format(err, "x") }
`,
		},
		{
			name: "single import converted to block",
			input: `package test
import "fmt"
func f(err error) error { return karma.Format(err, "x") }
`,
			want: `package test
import (
	"fmt"
	"github.com/reconquest/karma-go"
)
func f(err error) error { return karma.Format(err, "x") }
`,
		},
		{
			name: "aliased single import keeps its name",
			input: `package test
import f "fmt"
func f(err error) error { return karma.Format(err, "x") }
`,
			want: `package test
import (
	f "fmt"
	"github.com/reconquest/karma-go"
)
func f(err error) error { return karma.Format(err, "x") }
`,
		},
		{
			name: "added to existing import block",
			input: `package test
import (
	"fmt"
	"os"
)
func f(err error) error { return karma.Format(err, "x") }
`,
			want: `package test
import (
	"github.com/reconquest/karma-go"
	"fmt"
	"os"
)
func f(err error) error { return karma.Format(err, "x") }
`,
		},
		{
			name: "added to one-line import block",
			input: `package test
import ("fmt")
func f(err error) error { return karma.Format(err, "x") }
`,
			want: `package test
import (
	"github.com/reconquest/karma-go"
	"fmt")
func f(err error) error { return karma.Format(err, "x") }
`,
		},
		{
			name: "no import declaration",
			input: `package test
func f(err error) error { return karma.Format(err, "x") }
`,
			want: `package test

import "github.com/reconquest/karma-go"

func f(err error) error { return karma.Format(err, "x") }
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureKarmaImport(tt.input)
			if got != tt.want {
				t.Errorf("ensureKarmaImport() mismatch:\ngot:\n%s\n\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestQuoteMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		backtick bool
		want     string
	}{
		{"simple double-quoted", "connect", false, `"connect"`},
		{"simple backtick", "connect", true, "`connect`"},
		{"backtick with quote", `say "hi"`, true, "`say \"hi\"`"},
		{"backtick inside message", "a`b", true, `"a` + "`" + `b"`},
		{"quote escaped", `say "hi"`, false, `"say \"hi\""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteMessage(tt.message, tt.backtick)
			if got != tt.want {
				t.Errorf("quoteMessage(%q, %v) = %q, want %q", tt.message, tt.backtick, got, tt.want)
			}
		})
	}
}
