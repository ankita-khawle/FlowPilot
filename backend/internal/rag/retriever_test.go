package rag

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"flowpilot/backend/internal/ai"
)

type mockEmbeddingProvider struct {
	vectors [][]float32
	err     error
	inputs  []string
}

func (m *mockEmbeddingProvider) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	m.inputs = texts
	return m.vectors, m.err
}

type mockVectorStore struct {
	matches     []ChunkMatch
	err         error
	queryVector []float32
	documentID  string
	limit       int
	called      bool
}

func (m *mockVectorStore) UpsertChunks(
	ctx context.Context,
	chunks []Chunk,
	vectors [][]float32,
) error {
	return nil
}

func (m *mockVectorStore) Search(
	ctx context.Context,
	queryVector []float32,
	documentID string,
	limit int,
) ([]ChunkMatch, error) {
	m.called = true
	m.queryVector = queryVector
	m.documentID = documentID
	m.limit = limit
	return m.matches, m.err
}

func TestNewRAGRetriever(t *testing.T) {
	_, err := NewRAGRetriever(nil, &mockVectorStore{})
	if !errors.Is(err, ErrInvalidRetriever) {
		t.Fatalf("expected ErrInvalidRetriever, got %v", err)
	}

	_, err = NewRAGRetriever(&mockEmbeddingProvider{}, nil)
	if !errors.Is(err, ErrInvalidRetriever) {
		t.Fatalf("expected ErrInvalidRetriever, got %v", err)
	}
}

func TestRAGRetrieverRetrieve(t *testing.T) {
	embeddings := &mockEmbeddingProvider{
		vectors: [][]float32{{0.1, 0.2, 0.3}},
	}

	expectedMatches := []ChunkMatch{
		{
			Chunk: Chunk{
				ID:         "doc-123-chunk-0001",
				DocumentID: "doc-123",
				Index:      1,
				Text:       "Employee salary information",
				StartChar:  0,
				EndChar:    28,
			},
			Similarity: 0.95,
		},
	}

	store := &mockVectorStore{
		matches: expectedMatches,
	}

	retriever, err := NewRAGRetriever(embeddings, store)
	if err != nil {
		t.Fatal(err)
	}

	matches, err := retriever.Retrieve(
		context.Background(),
		"doc-123",
		"What is the employee salary?",
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(matches, expectedMatches) {
		t.Fatalf("unexpected matches: %+v", matches)
	}

	if !reflect.DeepEqual(
		embeddings.inputs,
		[]string{"What is the employee salary?"},
	) {
		t.Errorf("unexpected embedding input: %v", embeddings.inputs)
	}

	if !store.called {
		t.Fatal("expected vector store to be called")
	}

	if store.documentID != "doc-123" {
		t.Errorf("expected document ID doc-123, got %s", store.documentID)
	}

	if store.limit != 5 {
		t.Errorf("expected limit 5, got %d", store.limit)
	}

	if !reflect.DeepEqual(store.queryVector, embeddings.vectors[0]) {
		t.Error("query vector was not passed correctly")
	}
}

func TestRAGRetrieverRejectsInvalidInput(t *testing.T) {
	embeddings := &mockEmbeddingProvider{
		vectors: [][]float32{{0.1, 0.2}},
	}
	store := &mockVectorStore{}

	retriever, err := NewRAGRetriever(embeddings, store)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		documentID string
		question   string
		limit      int
		wantErr    error
	}{
		{"missing document ID", "", "question", 5, ErrEmptyDocumentID},
		{"blank question", "doc-123", "  ", 5, ErrEmptyQuestion},
		{"invalid limit", "doc-123", "question", 0, ErrInvalidSearchLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := retriever.Retrieve(
				context.Background(),
				tt.documentID,
				tt.question,
				tt.limit,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	if store.called {
		t.Error("vector store should not be called for invalid input")
	}
}

func TestRAGRetrieverHandlesEmbeddingError(t *testing.T) {
	expectedErr := errors.New("embedding provider unavailable")

	embeddings := &mockEmbeddingProvider{
		err: expectedErr,
	}
	store := &mockVectorStore{}

	retriever, _ := NewRAGRetriever(embeddings, store)

	_, err := retriever.Retrieve(
		context.Background(),
		"doc-123",
		"Find employee salary",
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected embedding error, got %v", err)
	}

	if store.called {
		t.Error("vector store should not be called after embedding failure")
	}
}

func TestRAGRetrieverRejectsUnexpectedEmbedding(t *testing.T) {
	tests := []struct {
		name    string
		vectors [][]float32
	}{
		{"no vectors", nil},
		{"multiple vectors", [][]float32{{0.1}, {0.2}}},
		{"empty vector", [][]float32{{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embeddings := &mockEmbeddingProvider{
				vectors: tt.vectors,
			}
			store := &mockVectorStore{}

			retriever, _ := NewRAGRetriever(embeddings, store)

			_, err := retriever.Retrieve(
				context.Background(),
				"doc-123",
				"Find employee salary",
				5,
			)

			if !errors.Is(err, ErrUnexpectedEmbedding) {
				t.Fatalf("expected ErrUnexpectedEmbedding, got %v", err)
			}

			if store.called {
				t.Error("vector store should not be called")
			}
		})
	}
}

func TestRAGRetrieverHandlesVectorStoreError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	embeddings := &mockEmbeddingProvider{
		vectors: [][]float32{{0.1, 0.2}},
	}
	store := &mockVectorStore{
		err: expectedErr,
	}

	retriever, _ := NewRAGRetriever(embeddings, store)

	_, err := retriever.Retrieve(
		context.Background(),
		"doc-123",
		"Find employee salary",
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected vector store error, got %v", err)
	}
}

var (
	_ ai.EmbeddingProvider = (*mockEmbeddingProvider)(nil)
	_ VectorStore          = (*mockVectorStore)(nil)
)
