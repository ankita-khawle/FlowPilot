package ai

import (
	"context"
	"errors"
)

var (
	ErrEmptyAPIKey      = errors.New("embedding API key is required")
	ErrEmptyInput       = errors.New("embedding input cannot be empty")
	ErrInvalidResponse  = errors.New("invalid embedding API response")
	ErrUnexpectedStatus = errors.New("unexpected embedding API status")
	ErrInvalidEmbedding = errors.New("invalid embedding vector")
)

// EmbeddingProvider defines a provider-independent embedding contract.
type EmbeddingProvider interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
