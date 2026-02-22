package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFix_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected /v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		var req chatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		if req.Model != "local-model" {
			t.Errorf("expected model local-model, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("expected system role, got %s", req.Messages[0].Role)
		}
		if req.Messages[0].Content != systemPrompt {
			t.Errorf("unexpected system prompt: %s", req.Messages[0].Content)
		}
		if req.Messages[1].Role != "user" {
			t.Errorf("expected user role, got %s", req.Messages[1].Role)
		}
		if req.Messages[1].Content != "# Helo Wrold\n\nThis is a tset." {
			t.Errorf("unexpected user content: %s", req.Messages[1].Content)
		}
		if req.Temperature != 0.3 {
			t.Errorf("expected temperature 0.3, got %f", req.Temperature)
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "# Hello World\n\nThis is a test."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithHTTP(server.URL, server.Client())
	result, err := client.Fix(context.Background(), "# Helo Wrold\n\nThis is a tset.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "# Hello World\n\nThis is a test." {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestFix_TrailingSlashInEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected /v1/chat/completions, got %s", r.URL.Path)
		}
		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "fixed"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithHTTP(server.URL+"/", server.Client())
	result, err := client.Fix(context.Background(), "broken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "fixed" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestFix_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewClientWithHTTP(server.URL, server.Client())
	_, err := client.Fix(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for server error response")
	}
	if want := "LMStudio returned status 500"; !contains(err.Error(), want) {
		t.Errorf("expected error to contain %q, got: %v", want, err)
	}
}

func TestFix_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClientWithHTTP(server.URL, server.Client())
	_, err := client.Fix(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for bad JSON response")
	}
	if want := "parsing response"; !contains(err.Error(), want) {
		t.Errorf("expected error to contain %q, got: %v", want, err)
	}
}

func TestFix_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{Choices: []chatChoice{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithHTTP(server.URL, server.Client())
	_, err := client.Fix(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for no choices")
	}
	if want := "no choices"; !contains(err.Error(), want) {
		t.Errorf("expected error to contain %q, got: %v", want, err)
	}
}

func TestFix_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: ""}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithHTTP(server.URL, server.Client())
	_, err := client.Fix(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if want := "empty content"; !contains(err.Error(), want) {
		t.Errorf("expected error to contain %q, got: %v", want, err)
	}
}

func TestFix_ConnectionRefused(t *testing.T) {
	client := NewClient("http://localhost:1")
	_, err := client.Fix(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
	if want := "sending request to LMStudio"; !contains(err.Error(), want) {
		t.Errorf("expected error to contain %q, got: %v", want, err)
	}
}

func TestFix_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server - the context will be canceled before this responds
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := NewClientWithHTTP(server.URL, server.Client())
	_, err := client.Fix(ctx, "test")
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	client := NewClient("http://localhost:1234")
	if client.endpoint != "http://localhost:1234" {
		t.Errorf("unexpected endpoint: %s", client.endpoint)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil HTTP client")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
