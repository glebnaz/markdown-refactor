package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	// Create the file first so WriteFile can stat it.
	if err := os.WriteFile(path, []byte("original content"), 0644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	err := WriteFile(path, "corrected content")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back file: %v", err)
	}

	if string(data) != "corrected content" {
		t.Errorf("expected 'corrected content', got %q", string(data))
	}
}

func TestWriteFile_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	err := WriteFile(path, "updated")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permissions 0600, got %v", info.Mode().Perm())
	}
}

func TestWriteFile_NonexistentFile(t *testing.T) {
	err := WriteFile("/nonexistent/path/file.md", "content")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	content, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if content != "hello world" {
		t.Errorf("expected 'hello world', got %q", content)
	}
}

func TestReadFile_NonexistentFile(t *testing.T) {
	_, err := ReadFile("/nonexistent/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestWriteFile_ThenReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.md")

	original := "# Heading\n\nSome text with *formatting*.\n"

	if err := os.WriteFile(path, []byte("placeholder"), 0644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	if err := WriteFile(path, original); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	content, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if content != original {
		t.Errorf("roundtrip mismatch:\nwant: %q\ngot:  %q", original, content)
	}
}

func TestWriteFile_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")

	if err := os.WriteFile(path, []byte("not empty"), 0644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	err := WriteFile(path, "")
	if err != nil {
		t.Fatalf("WriteFile with empty content: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back file: %v", err)
	}

	if string(data) != "" {
		t.Errorf("expected empty content, got %q", string(data))
	}
}
