package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func collectFiles(args []string) ([]string, error) {
	if len(args) == 0 {
		args = []string{"."}
	}

	var files []string
	var errs []error

	for _, arg := range args {
		expanded, err := expandPath(arg)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		files = append(files, expanded...)
	}

	return files, errors.Join(errs...)
}

func expandPath(path string) ([]string, error) {
	if root, ok := strings.CutSuffix(path, "/..."); ok {
		return walkGoFiles(root)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files := []string{}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}

	return files, nil
}

func walkGoFiles(root string) ([]string, error) {
	var files []string
	var errs []error

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		if info.IsDir() {
			if skippedDir(info.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}

	return files, errors.Join(errs...)
}

func skippedDir(name string) bool {
	return name == "vendor" ||
		strings.HasPrefix(name, ".") && name != "." && name != ".."
}
