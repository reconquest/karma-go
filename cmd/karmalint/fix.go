package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type Fix struct {
	Start       int
	End         int
	Replacement string
}

func fixFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	fixes, err := collectFixes(filename, content)
	if err != nil {
		return "", err
	}

	if len(fixes) == 0 {
		return string(content), nil
	}

	return finalizeFixes(filename, string(content), fixes), nil
}

func fixFileInPlace(filename string) (int, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	fixes, err := collectFixes(filename, content)
	if err != nil {
		return 0, err
	}

	if len(fixes) == 0 {
		return 0, nil
	}

	fixed := finalizeFixes(filename, string(content), fixes)

	err = os.WriteFile(filename, []byte(fixed), 0644) //nolint:gosec // G306: source files
	if err != nil {
		return 0, err
	}

	return len(fixes), nil
}

func collectFixes(filename string, content []byte) ([]Fix, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var fixes []Fix
	rewrite := mayRewriteErrorf(file)

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		if rewrite {
			if fix := collectFmtErrorfFix(call, content); fix != nil {
				fixes = append(fixes, *fix)
				return false
			}
		}

		_, msgIndex, ok := messageArgIndex(call)
		if !ok || msgIndex >= len(call.Args) {
			return true
		}

		arg := call.Args[msgIndex]
		message, ok := stringArg(arg)
		if !ok {
			return true
		}

		cleaned := cleanMessage(message)
		if cleaned == message {
			return true
		}

		literal := arg.(*ast.BasicLit)
		fixes = append(fixes, Fix{
			Start:       int(literal.Pos()) - 1,
			End:         int(literal.End()) - 1,
			Replacement: quoteMessage(cleaned, strings.HasPrefix(literal.Value, "`")),
		})

		return true
	})

	return fixes, nil
}

func collectFmtErrorfFix(call *ast.CallExpr, content []byte) *Fix {
	msg, literal, errName, ok := matchErrorfWrap(call)
	if !ok {
		return nil
	}

	cleaned := cleanMessage(trailingErrorPattern.ReplaceAllString(msg, ""))

	args := []string{errName, quoteMessage(cleaned, strings.HasPrefix(literal.Value, "`"))}

	for _, arg := range call.Args[1 : len(call.Args)-1] {
		source := getArgSource(arg, content)
		if source == "" {
			return nil
		}

		args = append(args, source)
	}

	return &Fix{
		Start:       int(call.Pos()) - 1,
		End:         int(call.End()) - 1,
		Replacement: "karma.Format(" + strings.Join(args, ", ") + ")",
	}
}

func finalizeFixes(filename, original string, fixes []Fix) string {
	fixed := applyFixes(original, fixes)
	fixed = ensureKarmaImport(fixed)
	return formatFixedFile(filename, original, fixed)
}

func applyFixes(content string, fixes []Fix) string {
	for i := range fixes {
		for j := i + 1; j < len(fixes); j++ {
			if fixes[i].Start < fixes[j].Start {
				fixes[i], fixes[j] = fixes[j], fixes[i]
			}
		}
	}

	result := content
	for _, fix := range fixes {
		result = result[:fix.Start] + fix.Replacement + result[fix.End:]
	}

	return result
}
