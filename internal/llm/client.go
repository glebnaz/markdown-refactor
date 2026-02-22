package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const systemPrompt = "Fix grammar, spelling, and syntax errors in the following Markdown. Do not change the meaning. Return only the corrected Markdown, no explanations."

// Client defines the interface for an LLM that can fix markdown text.
type Client interface {
	Fix(ctx context.Context, content string) (string, error)
}

// LMStudioClient communicates with an LMStudio server via the OpenAI-compatible API.
type LMStudioClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a new LMStudioClient with the given endpoint.
func NewClient(endpoint string) *LMStudioClient {
	return &LMStudioClient{
		endpoint:   strings.TrimRight(endpoint, "/"),
		httpClient: &http.Client{},
	}
}

// NewClientWithHTTP creates a new LMStudioClient with a custom http.Client (useful for testing).
func NewClientWithHTTP(endpoint string, httpClient *http.Client) *LMStudioClient {
	return &LMStudioClient{
		endpoint:   strings.TrimRight(endpoint, "/"),
		httpClient: httpClient,
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

// Fix sends the markdown content to the LLM and returns the corrected version.
func (c *LMStudioClient) Fix(ctx context.Context, content string) (string, error) {
	reqBody := chatRequest{
		Model: "local-model",
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: content},
		},
		Temperature: 0.3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	url := c.endpoint + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request to LMStudio: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LMStudio returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LMStudio returned no choices")
	}

	result := chatResp.Choices[0].Message.Content
	if result == "" {
		return "", fmt.Errorf("LMStudio returned empty content")
	}

	return result, nil
}
