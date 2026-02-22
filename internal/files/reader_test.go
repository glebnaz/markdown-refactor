package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSingleFile(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Hello"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}

	abs, _ := filepath.Abs(mdFile)
	if got[0] != abs {
		t.Errorf("expected %s, got %s", abs, got[0])
	}
}

func TestResolveMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{"a.md", "b.md", "c.markdown"}
	var paths []string

	for _, name := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}

	got, err := Resolve(paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d", len(got))
	}
}

func TestResolveDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create markdown files at root and nested level
	if err := os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "nested.md"), []byte("# Nested"), 0644); err != nil {
		t.Fatal(err)
	}
	// Non-markdown file should be ignored
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("text"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 markdown files, got %d: %v", len(got), got)
	}
}

func TestResolveDirectoryDeep(t *testing.T) {
	dir := t.TempDir()

	// Create deeply nested markdown file
	deep := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deep, "deep.md"), []byte("# Deep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "top.md"), []byte("# Top"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(got), got)
	}
}

func TestResolveMixedFilesAndDirs(t *testing.T) {
	dir := t.TempDir()

	// A standalone file
	standalone := filepath.Join(dir, "standalone.md")
	if err := os.WriteFile(standalone, []byte("# Standalone"), 0644); err != nil {
		t.Fatal(err)
	}

	// A directory with files
	subdir := filepath.Join(dir, "docs")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "doc1.md"), []byte("# Doc1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "doc2.md"), []byte("# Doc2"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve([]string{standalone, subdir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(got), got)
	}
}

func TestResolveDeduplicate(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Pass the same file twice
	got, err := Resolve([]string{mdFile, mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 file (deduplicated), got %d", len(got))
	}
}

func TestResolveMarkdownExtension(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "notes.markdown")
	if err := os.WriteFile(mdFile, []byte("# Notes"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 file, got %d", len(got))
	}
}

// Error cases

func TestResolveNoPaths(t *testing.T) {
	_, err := Resolve([]string{})
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
	if !contains(err.Error(), "no paths provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveFileNotFound(t *testing.T) {
	_, err := Resolve([]string{"/nonexistent/path/file.md"})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveNotMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	txtFile := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(txtFile, []byte("plain text"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Resolve([]string{txtFile})
	if err == nil {
		t.Fatal("expected error for non-markdown file")
	}
	if !contains(err.Error(), "not a markdown file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	_, err := Resolve([]string{dir})
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
	if !contains(err.Error(), "no markdown files found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveDirectoryWithNoMarkdown(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "code.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Resolve([]string{dir})
	if err == nil {
		t.Fatal("expected error for directory with no markdown")
	}
	if !contains(err.Error(), "no markdown files found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsMarkdown(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"file.md", true},
		{"file.MD", true},
		{"file.markdown", true},
		{"file.MARKDOWN", true},
		{"file.Md", true},
		{"file.txt", false},
		{"file.go", false},
		{"file", false},
		{"path/to/file.md", true},
	}

	for _, tt := range tests {
		got := isMarkdown(tt.path)
		if got != tt.want {
			t.Errorf("isMarkdown(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
