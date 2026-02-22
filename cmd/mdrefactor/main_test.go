package main

import (
	"testing"
)

func TestParseArgs_SingleFile(t *testing.T) {
	cfg, err := parseArgs([]string{"README.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.files) != 1 || cfg.files[0] != "README.md" {
		t.Errorf("expected files=[README.md], got %v", cfg.files)
	}
	if cfg.auto {
		t.Error("expected auto=false by default")
	}
	if cfg.endpoint != "http://localhost:1234" {
		t.Errorf("expected default endpoint, got %s", cfg.endpoint)
	}
}

func TestParseArgs_MultipleFiles(t *testing.T) {
	cfg, err := parseArgs([]string{"a.md", "b.md", "docs/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.files) != 3 {
		t.Errorf("expected 3 files, got %d", len(cfg.files))
	}
}

func TestParseArgs_AutoFlag(t *testing.T) {
	cfg, err := parseArgs([]string{"--auto", "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.auto {
		t.Error("expected auto=true")
	}
}

func TestParseArgs_EndpointFlag(t *testing.T) {
	cfg, err := parseArgs([]string{"--endpoint", "http://myhost:5000", "file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.endpoint != "http://myhost:5000" {
		t.Errorf("expected endpoint=http://myhost:5000, got %s", cfg.endpoint)
	}
}

func TestParseArgs_AllFlags(t *testing.T) {
	cfg, err := parseArgs([]string{"--auto", "--endpoint", "http://example:8080", "a.md", "b.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.auto {
		t.Error("expected auto=true")
	}
	if cfg.endpoint != "http://example:8080" {
		t.Errorf("expected endpoint=http://example:8080, got %s", cfg.endpoint)
	}
	if len(cfg.files) != 2 {
		t.Errorf("expected 2 files, got %d", len(cfg.files))
	}
}

func TestParseArgs_NoArgs(t *testing.T) {
	_, err := parseArgs([]string{})
	if err == nil {
		t.Error("expected error when no file arguments provided")
	}
}

func TestParseArgs_OnlyFlags(t *testing.T) {
	_, err := parseArgs([]string{"--auto"})
	if err == nil {
		t.Error("expected error when only flags and no file arguments provided")
	}
}

func TestParseArgs_InvalidFlag(t *testing.T) {
	_, err := parseArgs([]string{"--invalid", "file.md"})
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestParseArgs_DefaultEndpoint(t *testing.T) {
	cfg, err := parseArgs([]string{"file.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.endpoint != "http://localhost:1234" {
		t.Errorf("expected default endpoint http://localhost:1234, got %s", cfg.endpoint)
	}
}
