package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebnaz/markdown-refactor/internal/llm"
)

// mockLLMServer creates an HTTP test server that mimics LMStudio's OpenAI-compatible API.
// The corrector function takes input content and returns corrected content.
func mockLLMServer(t *testing.T, corrector func(string) string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// The user message is the last one.
		userContent := ""
		for _, m := range req.Messages {
			if m.Role == "user" {
				userContent = m.Content
			}
		}

		corrected := corrector(userContent)

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": corrected,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// createTempFiles creates temporary markdown files and returns their paths and a cleanup function.
func createTempFiles(t *testing.T, fileContents map[string]string) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range fileContents {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating dir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	return dir, func() {} // t.TempDir handles cleanup
}

func TestIntegration_AutoMode_SingleFile(t *testing.T) {
	server := mockLLMServer(t, func(content string) string {
		return strings.ReplaceAll(content, "teh", "the")
	})
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"test.md": "This is teh first line.\nAnd teh second line.\n",
	})

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{filepath.Join(dir, "test.md")},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Processing file 1/1: test.md") {
		t.Errorf("expected progress indicator, got: %s", output)
	}
	if !strings.Contains(output, "Applied changes to test.md") {
		t.Errorf("expected applied message, got: %s", output)
	}
	if !strings.Contains(output, "1 file(s) fixed, 0 skipped, 0 unchanged") {
		t.Errorf("expected summary, got: %s", output)
	}

	// Verify file was actually modified.
	data, err := os.ReadFile(filepath.Join(dir, "test.md"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "This is the first line.\nAnd the second line.\n" {
		t.Errorf("file content not corrected, got: %s", string(data))
	}
}

func TestIntegration_AutoMode_MultipleFiles(t *testing.T) {
	server := mockLLMServer(t, func(content string) string {
		return strings.ReplaceAll(content, "teh", "the")
	})
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"a.md": "Fix teh error.\n",
		"b.md": "Another teh mistake.\n",
	})

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Processing file 1/2: a.md") {
		t.Errorf("expected progress for a.md, got: %s", output)
	}
	if !strings.Contains(output, "Processing file 2/2: b.md") {
		t.Errorf("expected progress for b.md, got: %s", output)
	}
	if !strings.Contains(output, "2 file(s) fixed, 0 skipped, 0 unchanged") {
		t.Errorf("expected summary, got: %s", output)
	}

	// Verify both files were corrected.
	dataA, _ := os.ReadFile(filepath.Join(dir, "a.md"))
	if string(dataA) != "Fix the error.\n" {
		t.Errorf("a.md not corrected, got: %s", string(dataA))
	}
	dataB, _ := os.ReadFile(filepath.Join(dir, "b.md"))
	if string(dataB) != "Another the mistake.\n" {
		t.Errorf("b.md not corrected, got: %s", string(dataB))
	}
}

func TestIntegration_AutoMode_Directory(t *testing.T) {
	server := mockLLMServer(t, func(content string) string {
		return strings.ReplaceAll(content, "teh", "the")
	})
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"doc1.md":        "Fix teh first doc.\n",
		"subdir/doc2.md": "Fix teh second doc.\n",
		"skip.txt":       "Not a markdown file.\n",
	})

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{dir},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should process 2 markdown files, not the .txt file.
	if !strings.Contains(output, "2 file(s) fixed") {
		t.Errorf("expected 2 files fixed, got: %s", output)
	}
}

func TestIntegration_AutoMode_NoChanges(t *testing.T) {
	server := mockLLMServer(t, func(content string) string {
		return content // Return unchanged content.
	})
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"perfect.md": "This is already correct.\n",
	})

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{filepath.Join(dir, "perfect.md")},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No changes needed for perfect.md") {
		t.Errorf("expected no changes message, got: %s", output)
	}
	if !strings.Contains(output, "0 file(s) fixed, 0 skipped, 1 unchanged") {
		t.Errorf("expected summary with 1 unchanged, got: %s", output)
	}
}

func TestIntegration_AutoMode_MixedResults(t *testing.T) {
	server := mockLLMServer(t, func(content string) string {
		// Only fix content that contains "teh".
		return strings.ReplaceAll(content, "teh", "the")
	})
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"needs_fix.md": "Fix teh error.\n",
		"perfect.md":   "Already correct.\n",
	})

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{filepath.Join(dir, "needs_fix.md"), filepath.Join(dir, "perfect.md")},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1 file(s) fixed, 0 skipped, 1 unchanged") {
		t.Errorf("expected mixed summary, got: %s", output)
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	server := mockLLMServer(t, func(content string) string {
		return content
	})
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"test.md": "Some content.\n",
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{filepath.Join(dir, "test.md")},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(ctx, &buf, cfg, client)
	// The cancelled context should cause the file processing to return "quit" with context error.
	if err == nil {
		// Either it returned an error or the quit path was taken.
		output := buf.String()
		if !strings.Contains(output, "Quitting") {
			t.Errorf("expected quitting message or error, got output: %s", output)
		}
	}
}

func TestIntegration_LLMServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	dir, _ := createTempFiles(t, map[string]string{
		"test.md": "Some content.\n",
	})

	cfg := config{
		auto:     true,
		endpoint: server.URL,
		files:    []string{filepath.Join(dir, "test.md")},
	}

	client := llm.NewClient(server.URL)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err == nil {
		t.Error("expected error when LLM server returns 500")
	}
}

func TestIntegration_NonexistentFile(t *testing.T) {
	cfg := config{
		auto:     true,
		endpoint: "http://localhost:1234",
		files:    []string{"/nonexistent/file.md"},
	}

	client := llm.NewClient(cfg.endpoint)
	var buf bytes.Buffer
	err := execute(context.Background(), &buf, cfg, client)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
