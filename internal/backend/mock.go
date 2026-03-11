package backend

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockBackend is a simulated backend.
// It tokenizes the response and streams one token every 200ms to mimic
// incremental LLM generation.
// Error simulation via special input values:
//   - "error:timeout": simulate timeout (sleep 30s)
//   - "error:fail": simulate inference failure
type MockBackend struct{}

func NewMockBackend() *MockBackend {
	return &MockBackend{}
}

// StreamInfer simulates streaming inference.
// It tokenizes by spaces for ASCII text and by rune for non-ASCII text.
func (m *MockBackend) StreamInfer(ctx context.Context, input string, ch chan<- string) error {
	defer close(ch)

	// Error simulation triggered by special input values.
	if strings.HasPrefix(input, "error:timeout") {
		select {
		case <-time.After(30 * time.Second):
			return context.DeadlineExceeded
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if strings.HasPrefix(input, "error:fail") {
		return fmt.Errorf("simulated inference failure")
	}

	// Build synthetic reply content.
	reply := generateMockReply(input)
	tokens := tokenize(reply)

	// Stream one token every 200ms.
	for _, token := range tokens {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			ch <- token
		}
	}

	return nil
}

// generateMockReply builds a synthetic response for testing.
func generateMockReply(input string) string {
	return "This is a mock response to your input: " + input + ". Hope this helps!"
}

// tokenize splits text into streaming tokens.
// ASCII words are space-preserving tokens; non-ASCII words are rune-split.
func tokenize(text string) []string {
	var tokens []string
	words := strings.Fields(text)
	for i, word := range words {
		if isASCII(word) {
			if i > 0 {
				tokens = append(tokens, " "+word)
			} else {
				tokens = append(tokens, word)
			}
		} else {
			if i > 0 {
				tokens = append(tokens, " ")
			}
			for _, r := range word {
				tokens = append(tokens, string(r))
			}
		}
	}
	return tokens
}

// isASCII returns whether all bytes are ASCII.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}
