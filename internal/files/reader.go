package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolve takes a list of paths (files or directories) and returns a deduplicated
// list of existing, readable Markdown files. Directories are walked recursively.
func Resolve(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no paths provided")
	}

	seen := make(map[string]bool)
	var result []string

	for _, p := range paths {
		resolved, err := resolvePath(p)
		if err != nil {
			return nil, err
		}
		for _, f := range resolved {
			abs, err := filepath.Abs(f)
			if err != nil {
				return nil, fmt.Errorf("getting absolute path for %s: %w", f, err)
			}
			if !seen[abs] {
				seen[abs] = true
				result = append(result, abs)
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no markdown files found")
	}

	return result, nil
}

// resolvePath handles a single path: if it's a file, validates it; if it's a directory,
// walks it recursively looking for *.md files.
func resolvePath(p string) ([]string, error) {
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", p)
		}
		return nil, fmt.Errorf("accessing path %s: %w", p, err)
	}

	if info.IsDir() {
		return resolveDir(p)
	}

	if !isMarkdown(p) {
		return nil, fmt.Errorf("not a markdown file: %s", p)
	}

	return []string{p}, nil
}

// resolveDir walks a directory recursively and collects all *.md files.
func resolveDir(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isMarkdown(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", dir, err)
	}

	return files, nil
}

// isMarkdown checks if a file path has a markdown extension.
func isMarkdown(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}
