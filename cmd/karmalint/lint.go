package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const karmaImportPath = "github.com/reconquest/karma-go"

var (
	trailingErrorPattern   = regexp.MustCompile(`:\s*%[vsw]$`)
	errorfWrapPattern      = regexp.MustCompile(`:\s*%w$`)
	formatVerbPattern      = regexp.MustCompile(`%[+\-# 0]*(?:\d+)?(?:\.\d+)?[vTtbcdoOqxXUeEfFgGspw]`)
	trailingFailurePattern = regexp.MustCompile(`(?i)\s+(failed|error|err)$`)
)

var redundantPatterns = []struct {
	pattern     *regexp.Regexp
	description string
}{
	{regexp.MustCompile(`(?i)^failed to\s+`), `"failed to" is redundant`},
	{regexp.MustCompile(`(?i)^unable to\s+`), `"unable to" is redundant`},
	{regexp.MustCompile(`(?i)^could not\s+`), `"could not" is redundant`},
	{regexp.MustCompile(`(?i)^cannot\s+`), `"cannot" is redundant`},
	{regexp.MustCompile(`(?i)^can't\s+`), `"can't" is redundant`},
	{regexp.MustCompile(`(?i)^error\s+`), `"error" prefix is redundant`},
	{regexp.MustCompile(`(?i)^err:\s*`), `"err:" prefix is redundant`},
}

type Issue struct {
	Filename   string
	Line       int
	Column     int
	Message    string
	Original   string
	Suggestion string
	FuncName   string
}

func lintFile(filename string) ([]Issue, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	rewrite := mayRewriteErrorf(file)

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		if rewrite {
			if issue := checkFmtErrorfWrap(fset, call); issue != nil {
				issues = append(issues, *issue)
				return true
			}
		}

		funcName, msgIndex, ok := messageArgIndex(call)
		if !ok || msgIndex >= len(call.Args) {
			return true
		}

		message, ok := stringArg(call.Args[msgIndex])
		if !ok {
			return true
		}

		if issue := checkRedundantMessage(fset, call.Args[msgIndex], message, funcName); issue != nil {
			issues = append(issues, *issue)
		}

		return true
	})

	return issues, nil
}

func mayRewriteErrorf(file *ast.File) bool {
	if file.Name.Name == "karma" {
		return false
	}

	for _, spec := range file.Imports {
		if path, err := strconv.Unquote(spec.Path.Value); err == nil &&
			path == karmaImportPath {
			return spec.Name == nil
		}
	}

	return true
}

func checkFmtErrorfWrap(fset *token.FileSet, call *ast.CallExpr) *Issue {
	msg, _, errName, ok := matchErrorfWrap(call)
	if !ok {
		return nil
	}

	cleaned := cleanMessage(trailingErrorPattern.ReplaceAllString(msg, ""))
	quoted := strconv.Quote(cleaned)
	pos := fset.Position(call.Pos())

	message := "use karma.Format instead of fmt.Errorf for error wrapping"
	if !errorfWrapPattern.MatchString(msg) {
		message = "use karma.Format to keep the error chain instead of " +
			"formatting the error into the message"
	}

	suggestion := fmt.Sprintf("karma.Format(%s, %s)", errName, quoted)
	if args := formatArgs(call.Args[1 : len(call.Args)-1]); args != "" {
		suggestion = fmt.Sprintf("karma.Format(%s, %s, %s)", errName, quoted, args)
	}

	return &Issue{
		Filename:   pos.Filename,
		Line:       pos.Line,
		Column:     pos.Column,
		Message:    message,
		Original:   fmt.Sprintf("fmt.Errorf(%s, ...)", call.Args[0].(*ast.BasicLit).Value),
		Suggestion: suggestion,
		FuncName:   "fmt.Errorf",
	}
}

func matchErrorfWrap(call *ast.CallExpr) (string, *ast.BasicLit, string, bool) {
	if !isFmtErrorf(call) || len(call.Args) < 2 || call.Ellipsis.IsValid() {
		return "", nil, "", false
	}

	literal, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return "", nil, "", false
	}

	msg, ok := stringArg(literal)
	if !ok || !trailingErrorPattern.MatchString(msg) {
		return "", nil, "", false
	}

	unescaped := strings.ReplaceAll(msg, "%%", "")
	verbCount := len(formatVerbPattern.FindAllString(unescaped, -1))
	if verbCount != len(call.Args)-1 || verbCount < 1 {
		return "", nil, "", false
	}

	if strings.Count(unescaped, "%w") > 1 {
		return "", nil, "", false
	}

	errName := getArgName(call.Args[len(call.Args)-1])
	if errName == "" {
		return "", nil, "", false
	}

	shortName := errName[strings.LastIndex(errName, ".")+1:]
	if !looksLikeErrorName(shortName) {
		return "", nil, "", false
	}

	return msg, literal, errName, true
}

func isFmtErrorf(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "fmt" && sel.Sel.Name == "Errorf"
}

func checkRedundantMessage(
	fset *token.FileSet,
	arg ast.Expr,
	message string,
	funcName string,
) *Issue {
	for _, redundant := range redundantPatterns {
		if redundant.pattern.MatchString(message) {
			pos := fset.Position(arg.Pos())
			return &Issue{
				Filename:   pos.Filename,
				Line:       pos.Line,
				Column:     pos.Column,
				Message:    redundant.description,
				Original:   message,
				Suggestion: suggestFix(message, redundant.pattern),
				FuncName:   funcName,
			}
		}
	}

	if trailingFailurePattern.MatchString(message) {
		pos := fset.Position(arg.Pos())
		return &Issue{
			Filename:   pos.Filename,
			Line:       pos.Line,
			Column:     pos.Column,
			Message:    "trailing failure word is redundant",
			Original:   message,
			Suggestion: trailingFailurePattern.ReplaceAllString(message, ""),
			FuncName:   funcName,
		}
	}

	return nil
}

func messageArgIndex(call *ast.CallExpr) (string, int, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", 0, false
	}

	switch sel.Sel.Name {
	case "Format":
		if isKarmaRooted(sel.X) {
			return "karma.Format", 1, true
		}
	case "Collect":
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "karma" {
			return "karma.Collect", 0, true
		}
	}

	return "", 0, false
}

func isKarmaRooted(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name == "karma"
	case *ast.SelectorExpr:
		return isKarmaRooted(typed.X)
	case *ast.CallExpr:
		return isKarmaRooted(typed.Fun)
	}

	return false
}

func cleanMessage(message string) string {
	for _, redundant := range redundantPatterns {
		if redundant.pattern.MatchString(message) {
			message = suggestFix(message, redundant.pattern)
			break
		}
	}

	return trailingFailurePattern.ReplaceAllString(message, "")
}

func suggestFix(message string, pattern *regexp.Regexp) string {
	fixed := pattern.ReplaceAllString(message, "")

	first, size := utf8.DecodeRuneInString(fixed)
	if first == utf8.RuneError {
		return fixed
	}

	return strings.ToLower(string(first)) + fixed[size:]
}

func stringArg(arg ast.Expr) (string, bool) {
	literal, ok := arg.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}

	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		return "", false
	}

	return value, true
}

func looksLikeErrorName(name string) bool {
	return name == "err" || name == "error" ||
		strings.HasSuffix(name, "Err") || strings.HasSuffix(name, "Error")
}

func getArgName(arg ast.Expr) string {
	switch typed := arg.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		if ident, ok := typed.X.(*ast.Ident); ok {
			return ident.Name + "." + typed.Sel.Name
		}
	case *ast.BasicLit:
		return typed.Value
	}

	return ""
}

func getArgSource(arg ast.Expr, content []byte) string {
	if name := getArgName(arg); name != "" {
		return name
	}

	start := int(arg.Pos()) - 1
	end := int(arg.End()) - 1
	if start >= 0 && end <= len(content) && start < end {
		return string(content[start:end])
	}

	return ""
}

func formatArgs(args []ast.Expr) string {
	parts := []string{}

	for _, arg := range args {
		name := getArgName(arg)
		if name == "" {
			return "..."
		}

		parts = append(parts, name)
	}

	return strings.Join(parts, ", ")
}
