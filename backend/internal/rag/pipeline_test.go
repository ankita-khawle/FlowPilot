package rag

import (
	"context"
	"errors"
	"testing"

	"flowpilot/backend/internal/ai"
	"flowpilot/backend/internal/document"
)

type testDocumentExtractor struct {
	doc document.Document
	err error
}

func (e *testDocumentExtractor) Extract(
	ctx context.Context,
	fileName string,
	content []byte,
) (document.Document, error) {
	return e.doc, e.err
}

type testEmbeddingProvider struct {
	err   error
	calls int
}

func (e *testEmbeddingProvider) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	e.calls++

	if e.err != nil {
		return nil, e.err
	}

	vectors := make([][]float32, len(texts))
	for i := range texts {
		vectors[i] = []float32{0.1, 0.2, 0.3}
	}

	return vectors, nil
}

type testPipelineStore struct {
	chunks  []Chunk
	vectors [][]float32
	err     error
	calls   int
}

func (s *testPipelineStore) UpsertChunks(
	ctx context.Context,
	chunks []Chunk,
	vectors [][]float32,
) error {
	s.calls++
	s.chunks = chunks
	s.vectors = vectors
	return s.err
}

func (s *testPipelineStore) Search(
	ctx context.Context,
	queryVector []float32,
	documentID string,
	limit int,
) ([]ChunkMatch, error) {
	return nil, nil
}

type testAnswerer struct {
	response AnswerResponse
	err      error
}

func (a *testAnswerer) Answer(
	ctx context.Context,
	documentID string,
	question string,
	limit int,
) (AnswerResponse, error) {
	return a.response, a.err
}

var (
	_ ai.EmbeddingProvider = (*testEmbeddingProvider)(nil)
	_ VectorStore          = (*testPipelineStore)(nil)
	_ Answerer             = (*testAnswerer)(nil)
)

func TestNewPipelineRequiresDependencies(t *testing.T) {
	extractor := &testDocumentExtractor{}
	chunker, _ := NewTextChunker(100, 10)
	embeddings := &testEmbeddingProvider{}
	store := &testPipelineStore{}
	answerer := &testAnswerer{}

	tests := []struct {
		name      string
		extractor DocumentExtractor
		chunker   Chunker
		embedding ai.EmbeddingProvider
		store     VectorStore
		answerer  Answerer
	}{
		{"missing extractor", nil, chunker, embeddings, store, answerer},
		{"missing chunker", extractor, nil, embeddings, store, answerer},
		{"missing embedding provider", extractor, chunker, nil, store, answerer},
		{"missing vector store", extractor, chunker, embeddings, nil, answerer},
		{"missing answerer", extractor, chunker, embeddings, store, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPipeline(
				tt.extractor,
				tt.chunker,
				tt.embedding,
				tt.store,
				tt.answerer,
			)

			if err == nil {
				t.Fatal("expected dependency validation error")
			}
		})
	}
}

func TestPipelineIngest(t *testing.T) {
	doc := document.Document{
		ID:             "doc-123",
		FileName:       "employee.txt",
		FileType:       "txt",
		Status:         "completed",
		Text:           "Employee name is John. Annual salary is INR 25 lakh.",
		WordCount:      9,
		CharacterCount: 54,
		Pages: []document.PageContent{
			{
				Text: "Employee name is John. Annual salary is INR 25 lakh.",
			},
		},
	}

	extractor := &testDocumentExtractor{doc: doc}
	chunker, err := NewTextChunker(100, 10)
	if err != nil {
		t.Fatal(err)
	}

	embeddings := &testEmbeddingProvider{}
	store := &testPipelineStore{}

	pipeline, err := NewPipeline(
		extractor,
		chunker,
		embeddings,
		store,
		&testAnswerer{},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := pipeline.Ingest(
		context.Background(),
		"employee.txt",
		[]byte(doc.Text),
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.DocumentID != doc.ID {
		t.Errorf("expected document ID %s, got %s", doc.ID, result.DocumentID)
	}

	if result.ChunkCount != 1 {
		t.Errorf("expected one chunk, got %d", result.ChunkCount)
	}

	if store.calls != 1 {
		t.Errorf("expected one storage call, got %d", store.calls)
	}

	if len(store.chunks) != len(store.vectors) {
		t.Error("chunk and vector counts do not match")
	}

	if embeddings.calls != 1 {
		t.Errorf("expected one embedding call, got %d", embeddings.calls)
	}
}

func TestPipelineIngestEmbeddingFailure(t *testing.T) {
	expectedErr := errors.New("embedding provider unavailable")

	doc := document.Document{
		ID:   "doc-123",
		Text: "Some document content that needs embedding.",
	}

	extractor := &testDocumentExtractor{doc: doc}
	chunker, _ := NewTextChunker(100, 10)
	embeddings := &testEmbeddingProvider{err: expectedErr}
	store := &testPipelineStore{}

	pipeline, _ := NewPipeline(
		extractor,
		chunker,
		embeddings,
		store,
		&testAnswerer{},
	)

	_, err := pipeline.Ingest(
		context.Background(),
		"document.txt",
		[]byte(doc.Text),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected embedding error, got %v", err)
	}

	if store.calls != 0 {
		t.Error("storage should not be called when embedding fails")
	}
}

func TestPipelineIngestStorageFailure(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	doc := document.Document{
		ID:   "doc-123",
		Text: "Some document content that needs storing.",
	}

	extractor := &testDocumentExtractor{doc: doc}
	chunker, _ := NewTextChunker(100, 10)
	embeddings := &testEmbeddingProvider{}
	store := &testPipelineStore{err: expectedErr}

	pipeline, _ := NewPipeline(
		extractor,
		chunker,
		embeddings,
		store,
		&testAnswerer{},
	)

	_, err := pipeline.Ingest(
		context.Background(),
		"document.txt",
		[]byte(doc.Text),
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected storage error, got %v", err)
	}
}

func TestPipelineAsk(t *testing.T) {
	expected := AnswerResponse{
		Answer: "The annual salary is INR 25 lakh.",
	}

	answerer := &testAnswerer{response: expected}

	pipeline, err := NewPipeline(
		&testDocumentExtractor{},
		&TextChunker{},
		&testEmbeddingProvider{},
		&testPipelineStore{},
		answerer,
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := pipeline.Ask(
		context.Background(),
		"doc-123",
		"What is the salary?",
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Answer != expected.Answer {
		t.Errorf("expected %q, got %q", expected.Answer, result.Answer)
	}
}
