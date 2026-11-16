package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewOpenAIEmbeddingProvider(t *testing.T) {
	_, err := NewOpenAIEmbeddingProvider("")
	if !errors.Is(err, ErrEmptyAPIKey) {
		t.Fatalf("expected ErrEmptyAPIKey, got %v", err)
	}

	_, err = NewOpenAIEmbeddingProvider(
		"test-key",
		WithEmbeddingDimensions(-1),
	)
	if err == nil {
		t.Fatal("expected error for negative dimensions")
	}

	provider, err := NewOpenAIEmbeddingProvider("test-key")
	if err != nil {
		t.Fatal(err)
	}

	if provider.model != "text-embedding-3-small" {
		t.Errorf("unexpected model: %s", provider.model)
	}
}

func TestOpenAIEmbeddingProviderEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Error("missing or incorrect authorization header")
			}

			var request openAIEmbeddingRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}

			if request.Model != "text-embedding-3-small" {
				t.Errorf("unexpected model: %s", request.Model)
			}

			if len(request.Input) != 2 {
				t.Fatalf("expected two input texts, got %d", len(request.Input))
			}

			// Deliberately return vectors out of order.
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(openAIEmbeddingResponse{
				Data: []openAIEmbeddingData{
					{Index: 1, Embedding: []float32{0.4, 0.5, 0.6}},
					{Index: 0, Embedding: []float32{0.1, 0.2, 0.3}},
				},
			})
		},
	))
	defer server.Close()

	provider, err := NewOpenAIEmbeddingProvider(
		"test-key",
		WithEmbeddingBaseURL(server.URL),
	)
	if err != nil {
		t.Fatal(err)
	}

	vectors, err := provider.Embed(
		context.Background(),
		[]string{"First document chunk", "Second document chunk"},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := [][]float32{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	if !reflect.DeepEqual(vectors, expected) {
		t.Fatalf("expected %v, got %v", expected, vectors)
	}
}

func TestOpenAIEmbeddingProviderEmptyInput(t *testing.T) {
	provider, _ := NewOpenAIEmbeddingProvider("test-key")

	vectors, err := provider.Embed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(vectors) != 0 {
		t.Fatalf("expected empty result, got %v", vectors)
	}
}

func TestOpenAIEmbeddingProviderRejectsBlankInput(t *testing.T) {
	provider, _ := NewOpenAIEmbeddingProvider("test-key")

	_, err := provider.Embed(
		context.Background(),
		[]string{"valid text", "   "},
	)

	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}

func TestOpenAIEmbeddingProviderHandlesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(
				`{"error":{"message":"Invalid API key"}}`,
			))
		},
	))
	defer server.Close()

	provider, _ := NewOpenAIEmbeddingProvider(
		"invalid-key",
		WithEmbeddingBaseURL(server.URL),
	)

	_, err := provider.Embed(
		context.Background(),
		[]string{"some text"},
	)

	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf("expected ErrUnexpectedStatus, got %v", err)
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Fatalf("expected API error message, got %v", err)
	}
}

func TestOpenAIEmbeddingProviderRejectsInvalidVector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(openAIEmbeddingResponse{
				Data: []openAIEmbeddingData{
					{Index: 0, Embedding: []float32{}},
				},
			})
		},
	))
	defer server.Close()

	provider, _ := NewOpenAIEmbeddingProvider(
		"test-key",
		WithEmbeddingBaseURL(server.URL),
	)

	_, err := provider.Embed(
		context.Background(),
		[]string{"some text"},
	)

	if !errors.Is(err, ErrInvalidEmbedding) {
		t.Fatalf("expected ErrInvalidEmbedding, got %v", err)
	}
}

func TestOpenAIEmbeddingProviderRejectsMissingVector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(openAIEmbeddingResponse{
				Data: []openAIEmbeddingData{},
			})
		},
	))
	defer server.Close()

	provider, _ := NewOpenAIEmbeddingProvider(
		"test-key",
		WithEmbeddingBaseURL(server.URL),
	)

	_, err := provider.Embed(
		context.Background(),
		[]string{"some text"},
	)

	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("expected ErrInvalidResponse, got %v", err)
	}
}
