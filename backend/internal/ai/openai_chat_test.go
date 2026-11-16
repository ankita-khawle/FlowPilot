package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewOpenAIChatProvider(t *testing.T) {
	_, err := NewOpenAIChatProvider("")
	if !errors.Is(err, ErrEmptyChatAPIKey) {
		t.Fatalf("expected ErrEmptyChatAPIKey, got %v", err)
	}

	_, err = NewOpenAIChatProvider(
		"test-key",
		WithChatModel(""),
	)
	if err == nil {
		t.Fatal("expected error for empty model")
	}

	provider, err := NewOpenAIChatProvider("test-key")
	if err != nil {
		t.Fatal(err)
	}

	if provider.model != "gpt-4o-mini" {
		t.Errorf("unexpected model: %s", provider.model)
	}
}

func TestOpenAIChatProviderComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Error("incorrect authorization header")
			}

			var request openAIChatRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}

			if request.Model != "gpt-4o-mini" {
				t.Errorf("unexpected model: %s", request.Model)
			}

			if request.ResponseFormat.Type != "json_object" {
				t.Errorf("expected JSON response format, got %s",
					request.ResponseFormat.Type)
			}

			if len(request.Messages) != 1 {
				t.Fatalf("expected one message, got %d", len(request.Messages))
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []any{
					map[string]any{
						"message": map[string]string{
							"role":    "assistant",
							"content": `{"answer":"Hello","source_ids":[]}`,
						},
					},
				},
			})
		},
	))
	defer server.Close()

	provider, err := NewOpenAIChatProvider(
		"test-key",
		WithChatBaseURL(server.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	output, err := provider.Complete(
		context.Background(),
		ChatCompletionRequest{
			Messages: []ChatMessage{
				{Role: "user", Content: "Hello"},
			},
			Temperature: 0.1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if output != `{"answer":"Hello","source_ids":[]}` {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestOpenAIChatProviderRejectsEmptyMessages(t *testing.T) {
	provider, _ := NewOpenAIChatProvider("test-key")

	_, err := provider.Complete(
		context.Background(),
		ChatCompletionRequest{},
	)

	if !errors.Is(err, ErrEmptyMessages) {
		t.Fatalf("expected ErrEmptyMessages, got %v", err)
	}
}

func TestOpenAIChatProviderHandlesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(
				`{"error":{"message":"Invalid API key"}}`,
			))
		},
	))
	defer server.Close()

	provider, _ := NewOpenAIChatProvider(
		"invalid-key",
		WithChatBaseURL(server.URL),
	)

	_, err := provider.Complete(
		context.Background(),
		ChatCompletionRequest{
			Messages: []ChatMessage{
				{Role: "user", Content: "Hello"},
			},
		},
	)

	if err == nil || !strings.Contains(err.Error(), "Invalid API key") {
		t.Fatalf("expected API error message, got %v", err)
	}
}

func TestOpenAIChatProviderRejectsEmptyCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"choices":[]}`))
		},
	))
	defer server.Close()

	provider, _ := NewOpenAIChatProvider(
		"test-key",
		WithChatBaseURL(server.URL),
	)

	_, err := provider.Complete(
		context.Background(),
		ChatCompletionRequest{
			Messages: []ChatMessage{
				{Role: "user", Content: "Hello"},
			},
		},
	)

	if !errors.Is(err, ErrEmptyCompletion) {
		t.Fatalf("expected ErrEmptyCompletion, got %v", err)
	}
}
