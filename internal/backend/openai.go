package backend

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// OpenAIBackend integrates with the OpenAI Chat Completions API in stream mode.
// Required environment variable: OPENAI_API_KEY.
// Optional environment variables: OPENAI_BASE_URL, OPENAI_MODEL.
type OpenAIBackend struct {
	apiKey  string
	baseURL string
	model   string
}

func NewOpenAIBackend() *OpenAIBackend {
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}
	return &OpenAIBackend{
		apiKey:  os.Getenv("OPENAI_API_KEY"),
		baseURL: baseURL,
		model:   modelName,
	}
}

// openAIRequest is the OpenAI API request payload.
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIStreamChunk represents one chunk in the SSE stream.
type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// StreamInfer calls OpenAI Chat Completions (stream=true) and forwards tokens.
func (o *OpenAIBackend) StreamInfer(ctx context.Context, input string, ch chan<- string) error {
	defer close(ch)

	if o.apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Build request payload.
	reqBody := openAIRequest{
		Model: o.model,
		Messages: []openAIMessage{
			{Role: "user", Content: input},
		},
		Stream: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build HTTP request.
	url := o.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	// Send request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to OpenAI failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream response.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE lines start with "data: ".
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		// End-of-stream marker.
		if data == "[DONE]" {
			break
		}

		// Parse JSON chunk.
		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Extract token content and forward.
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- chunk.Choices[0].Delta.Content:
			}
		}
	}

	return scanner.Err()
}
