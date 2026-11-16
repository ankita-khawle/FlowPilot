package rag

import (
	"context"
	"errors"
)

var (
	ErrInvalidVector       = errors.New("invalid embedding vector")
	ErrVectorCountMismatch = errors.New("chunk and vector counts do not match")
	ErrInvalidSearchLimit  = errors.New("search limit must be greater than zero")
	ErrEmptyDocumentID     = errors.New("document ID is required for retrieval")
)

type ChunkMatch struct {
	Chunk      Chunk   `json:"chunk"`
	Similarity float64 `json:"similarity"`
}

type VectorStore interface {
	UpsertChunks(
		ctx context.Context,
		chunks []Chunk,
		vectors [][]float32,
	) error

	Search(
		ctx context.Context,
		queryVector []float32,
		documentID string,
		limit int,
	) ([]ChunkMatch, error)
}
