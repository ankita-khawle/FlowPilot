package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"flowpilot/backend/internal/rag"
)

type mockRAGPipeline struct {
	ingestResult rag.IngestResult
	answerResult rag.AnswerResponse
	ingestErr    error
	answerErr    error

	ingestCalled bool
	askCalled    bool
	documentID   string
	question     string
	limit        int
}

func (m *mockRAGPipeline) Ingest(
	ctx context.Context,
	fileName string,
	content []byte,
) (rag.IngestResult, error) {
	m.ingestCalled = true
	return m.ingestResult, m.ingestErr
}

func (m *mockRAGPipeline) Ask(
	ctx context.Context,
	documentID string,
	question string,
	limit int,
) (rag.AnswerResponse, error) {
	m.askCalled = true
	m.documentID = documentID
	m.question = question
	m.limit = limit
	return m.answerResult, m.answerErr
}

var _ RAGPipeline = (*mockRAGPipeline)(nil)

func TestRAGAskHandler(t *testing.T) {
	pipeline := &mockRAGPipeline{
		answerResult: rag.AnswerResponse{
			Answer: "The salary is INR 25 lakh.",
			Citations: []rag.Citation{
				{
					DocumentID: "doc-123",
					ChunkID:    "chunk-001",
					Excerpt:    "Annual salary is INR 25 lakh.",
				},
			},
		},
	}

	handler := &Handler{ragPipeline: pipeline}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/rag/documents/doc-123/ask",
		strings.NewReader(`{"question":"What is the salary?"}`),
	)
	request.SetPathValue("documentID", "doc-123")

	recorder := httptest.NewRecorder()
	handler.ragAsk(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var result rag.AnswerResponse
	if err := json.NewDecoder(recorder.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result.Answer != pipeline.answerResult.Answer {
		t.Errorf("unexpected answer: %s", result.Answer)
	}

	if !pipeline.askCalled {
		t.Fatal("expected pipeline Ask to be called")
	}

	if pipeline.documentID != "doc-123" {
		t.Errorf("unexpected document ID: %s", pipeline.documentID)
	}

	if pipeline.question != "What is the salary?" {
		t.Errorf("unexpected question: %s", pipeline.question)
	}

	if pipeline.limit != defaultRetrievalLimit {
		t.Errorf("expected default limit %d, got %d",
			defaultRetrievalLimit, pipeline.limit)
	}
}

func TestRAGAskHandlerRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name       string
		documentID string
		body       string
	}{
		{"missing document ID", "", `{"question":"Hello"}`},
		{"missing question", "doc-123", `{}`},
		{"invalid JSON", "doc-123", `not-json`},
		{"unknown field", "doc-123", `{"question":"Hello","extra":true}`},
		{"invalid limit", "doc-123", `{"question":"Hello","limit":100}`},
		{"blank question", "doc-123", `{"question":"  "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &mockRAGPipeline{}
			handler := &Handler{ragPipeline: pipeline}

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/rag/documents/"+tt.documentID+"/ask",
				strings.NewReader(tt.body),
			)
			request.SetPathValue("documentID", tt.documentID)

			recorder := httptest.NewRecorder()
			handler.ragAsk(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}

			if pipeline.askCalled {
				t.Error("pipeline should not be called for invalid requests")
			}
		})
	}
}

func TestRAGAskHandlerUnavailable(t *testing.T) {
	handler := &Handler{}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/rag/documents/doc-123/ask",
		strings.NewReader(`{"question":"Hello"}`),
	)
	request.SetPathValue("documentID", "doc-123")

	recorder := httptest.NewRecorder()
	handler.ragAsk(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestIsSupportedDocument(t *testing.T) {
    tests := []struct {
        name      string
        extension string
        want      bool
    }{
        {"PDF", ".pdf", true},
        {"DOCX", ".docx", true},
        {"TXT", ".txt", true},
        {"uppercase TXT", ".TXT", true},
        {"unsupported format", ".exe", false},
        {"empty extension", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := isSupportedDocument(tt.extension)
            if got != tt.want {
                t.Errorf("isSupportedDocument(%q) = %v, want %v",
                    tt.extension, got, tt.want)
            }
        })
    }
}