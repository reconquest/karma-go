package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"
)

func formatFixedFile(filename, original, fixed string) string {
	formatted, err := format.Source([]byte(original))
	if err != nil ||
		strings.TrimRight(string(formatted), "\n") != strings.TrimRight(original, "\n") {
		return fixed
	}

	processed, err := imports.Process(filename, []byte(fixed), nil)
	if err != nil {
		return fixed
	}

	return string(processed)
}

func ensureKarmaImport(content string) string {
	if !strings.Contains(content, "karma.") {
		return content
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixed.go", content, parser.ParseComments)
	if err != nil {
		return content
	}

	for _, spec := range file.Imports {
		if path, err := strconv.Unquote(spec.Path.Value); err == nil &&
			path == karmaImportPath {
			return content
		}
	}

	quoted := strconv.Quote(karmaImportPath)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}

		if gen.Lparen.IsValid() {
			insertAt := fset.Position(gen.Lparen).Offset + 1
			if insertAt < len(content) && content[insertAt] != '\n' {
				return content[:insertAt] + "\n\t" + quoted + "\n\t" + content[insertAt:]
			}

			return content[:insertAt] + "\n\t" + quoted + content[insertAt:]
		}

		if len(gen.Specs) != 1 {
			continue
		}

		spec, ok := gen.Specs[0].(*ast.ImportSpec)
		if !ok {
			continue
		}

		name := ""
		if spec.Name != nil {
			name = spec.Name.Name + " "
		}

		start := fset.Position(gen.Pos()).Offset
		end := fset.Position(gen.End()).Offset
		block := "import (\n\t" + name + spec.Path.Value + "\n\t" + quoted + "\n)"

		return content[:start] + block + content[end:]
	}

	insertAt := fset.Position(file.Name.End()).Offset
	return content[:insertAt] + "\n\nimport " + quoted + "\n" + content[insertAt:]
}

func quoteMessage(message string, backtick bool) string {
	if backtick && !strings.Contains(message, "`") {
		return "`" + message + "`"
	}

	return strconv.Quote(message)
}
