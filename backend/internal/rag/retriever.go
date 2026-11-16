package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flowpilot/backend/internal/ai"
)

var (
	ErrEmptyQuestion       = errors.New("question cannot be empty")
	ErrInvalidRetriever    = errors.New("embedding provider and vector store are required")
	ErrUnexpectedEmbedding = errors.New("embedding provider returned an unexpected result")
)

type Retriever interface {
	Retrieve(
		ctx context.Context,
		documentID string,
		question string,
		limit int,
	) ([]ChunkMatch, error)
}

type RAGRetriever struct {
	embeddings ai.EmbeddingProvider
	store      VectorStore
}

func NewRAGRetriever(
	embeddings ai.EmbeddingProvider,
	store VectorStore,
) (*RAGRetriever, error) {
	if embeddings == nil || store == nil {
		return nil, ErrInvalidRetriever
	}

	return &RAGRetriever{
		embeddings: embeddings,
		store:      store,
	}, nil
}

func (r *RAGRetriever) Retrieve(
	ctx context.Context,
	documentID string,
	question string,
	limit int,
) ([]ChunkMatch, error) {
	if r == nil || r.embeddings == nil || r.store == nil {
		return nil, ErrInvalidRetriever
	}

	if strings.TrimSpace(documentID) == "" {
		return nil, ErrEmptyDocumentID
	}

	if strings.TrimSpace(question) == "" {
		return nil, ErrEmptyQuestion
	}

	if limit <= 0 {
		return nil, ErrInvalidSearchLimit
	}

	// Convert the user's question into an embedding.
	vectors, err := r.embeddings.Embed(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("embed retrieval question: %w", err)
	}

	if len(vectors) != 1 || len(vectors[0]) == 0 {
		return nil, ErrUnexpectedEmbedding
	}

	// Search only within the requested document.
	matches, err := r.store.Search(
		ctx,
		vectors[0],
		documentID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("retrieve document chunks: %w", err)
	}

	return matches, nil
}
