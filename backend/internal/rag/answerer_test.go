package rag

import (
	"context"
	"errors"
	"strings"
	"testing"

	"flowpilot/backend/internal/ai"
)

type mockRetriever struct {
	matches    []ChunkMatch
	err        error
	called     bool
	documentID string
	question   string
	limit      int
}

func (m *mockRetriever) Retrieve(
	ctx context.Context,
	documentID string,
	question string,
	limit int,
) ([]ChunkMatch, error) {
	m.called = true
	m.documentID = documentID
	m.question = question
	m.limit = limit
	return m.matches, m.err
}

type mockLLMProvider struct {
	output  string
	err     error
	request ai.ChatCompletionRequest
	called  bool
}

func (m *mockLLMProvider) Complete(
	ctx context.Context,
	request ai.ChatCompletionRequest,
) (string, error) {
	m.called = true
	m.request = request
	return m.output, m.err
}

func TestNewRAGAnswerer(t *testing.T) {
	_, err := NewRAGAnswerer(nil, &mockLLMProvider{})
	if !errors.Is(err, ErrInvalidAnswerer) {
		t.Fatalf("expected ErrInvalidAnswerer, got %v", err)
	}

	_, err = NewRAGAnswerer(&mockRetriever{}, nil)
	if !errors.Is(err, ErrInvalidAnswerer) {
		t.Fatalf("expected ErrInvalidAnswerer, got %v", err)
	}
}

func TestRAGAnswererAnswer(t *testing.T) {
	page := 2

	retriever := &mockRetriever{
		matches: []ChunkMatch{
			{
				Chunk: Chunk{
					ID:         "chunk-001",
					DocumentID: "doc-123",
					Index:      1,
					Text:       "The employee's annual salary is INR 25 lakh.",
					PageNumber: &page,
					StartChar:  100,
					EndChar:    146,
				},
				Similarity: 0.94,
			},
		},
	}

	llm := &mockLLMProvider{
		output: `{
			"answer": "The annual salary is INR 25 lakh.",
			"source_ids": ["chunk-001"]
		}`,
	}

	answerer, err := NewRAGAnswerer(retriever, llm)
	if err != nil {
		t.Fatal(err)
	}

	result, err := answerer.Answer(
		context.Background(),
		"doc-123",
		"What is the annual salary?",
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Answer != "The annual salary is INR 25 lakh." {
		t.Errorf("unexpected answer: %s", result.Answer)
	}

	if len(result.Citations) != 1 {
		t.Fatalf("expected one citation, got %d", len(result.Citations))
	}

	citation := result.Citations[0]

	if citation.ChunkID != "chunk-001" {
		t.Errorf("unexpected chunk ID: %s", citation.ChunkID)
	}

	if citation.DocumentID != "doc-123" {
		t.Errorf("unexpected document ID: %s", citation.DocumentID)
	}

	if citation.PageNumber == nil || *citation.PageNumber != 2 {
		t.Error("expected citation to preserve page number")
	}

	if citation.StartChar != 100 || citation.EndChar != 146 {
		t.Error("unexpected citation offsets")
	}

	if !retriever.called || !llm.called {
		t.Error("expected both retriever and LLM to be called")
	}
}

func TestRAGAnswererRejectsUnknownSources(t *testing.T) {
	retriever := &mockRetriever{
		matches: []ChunkMatch{
			{
				Chunk: Chunk{
					ID:         "valid-chunk",
					DocumentID: "doc-123",
					Text:       "The company has 100 employees.",
				},
			},
		},
	}

	llm := &mockLLMProvider{
		output: `{
			"answer": "The company has 100 employees.",
			"source_ids": ["valid-chunk", "fabricated-chunk"]
		}`,
	}

	answerer, _ := NewRAGAnswerer(retriever, llm)

	result, err := answerer.Answer(
		context.Background(),
		"doc-123",
		"How many employees?",
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Citations) != 1 {
		t.Fatalf("expected only one valid citation, got %d", len(result.Citations))
	}

	if result.Citations[0].ChunkID != "valid-chunk" {
		t.Errorf("unexpected citation: %+v", result.Citations[0])
	}
}

func TestRAGAnswererDeduplicatesSources(t *testing.T) {
	retriever := &mockRetriever{
		matches: []ChunkMatch{
			{
				Chunk: Chunk{
					ID:         "chunk-001",
					DocumentID: "doc-123",
					Text:       "Some document content.",
				},
			},
		},
	}

	llm := &mockLLMProvider{
		output: `{
			"answer": "Some document content.",
			"source_ids": ["chunk-001", "chunk-001"]
		}`,
	}

	answerer, _ := NewRAGAnswerer(retriever, llm)

	result, err := answerer.Answer(
		context.Background(),
		"doc-123",
		"Summarize this content",
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Citations) != 1 {
		t.Fatalf("expected one deduplicated citation, got %d", len(result.Citations))
	}
}

func TestRAGAnswererHandlesNoMatches(t *testing.T) {
	retriever := &mockRetriever{
		matches: []ChunkMatch{},
	}
	llm := &mockLLMProvider{}

	answerer, _ := NewRAGAnswerer(retriever, llm)

	result, err := answerer.Answer(
		context.Background(),
		"doc-123",
		"Who is the CEO?",
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Citations) != 0 {
		t.Error("expected no citations")
	}

	if llm.called {
		t.Error("LLM should not be called when there are no matches")
	}

	if !strings.Contains(result.Answer, "couldn't find relevant information") {
		t.Errorf("unexpected fallback answer: %s", result.Answer)
	}
}

func TestRAGAnswererHandlesInvalidLLMOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"invalid JSON", "not JSON"},
		{"empty answer", `{"answer":"  ","source_ids":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retriever := &mockRetriever{
				matches: []ChunkMatch{
					{
						Chunk: Chunk{
							ID:   "chunk-001",
							Text: "Document content",
						},
					},
				},
			}

			llm := &mockLLMProvider{output: tt.output}
			answerer, _ := NewRAGAnswerer(retriever, llm)

			_, err := answerer.Answer(
				context.Background(),
				"doc-123",
				"Question",
				5,
			)

			if !errors.Is(err, ErrInvalidLLMOutput) {
				t.Fatalf("expected ErrInvalidLLMOutput, got %v", err)
			}
		})
	}
}

func TestRAGAnswererHandlesLLMError(t *testing.T) {
	expectedErr := errors.New("provider unavailable")

	retriever := &mockRetriever{
		matches: []ChunkMatch{
			{
				Chunk: Chunk{
					ID:   "chunk-001",
					Text: "Document content",
				},
			},
		},
	}

	llm := &mockLLMProvider{err: expectedErr}
	answerer, _ := NewRAGAnswerer(retriever, llm)

	_, err := answerer.Answer(
		context.Background(),
		"doc-123",
		"Question",
		5,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestRAGAnswererRejectsBlankQuestion(t *testing.T) {
	retriever := &mockRetriever{}
	llm := &mockLLMProvider{}
	answerer, _ := NewRAGAnswerer(retriever, llm)

	_, err := answerer.Answer(
		context.Background(),
		"doc-123",
		"   ",
		5,
	)

	if !errors.Is(err, ErrEmptyQuestion) {
		t.Fatalf("expected ErrEmptyQuestion, got %v", err)
	}

	if retriever.called || llm.called {
		t.Error("providers should not be called for a blank question")
	}
}
