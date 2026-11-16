package rag

import (
	"errors"
	"strings"
	"testing"

	"flowpilot/backend/internal/document"
)

func TestNewTextChunker(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		overlap int
		wantErr error
	}{
		{"valid config", 100, 20, nil},
		{"zero chunk size", 0, 10, ErrInvalidChunkSize},
		{"negative chunk size", -1, 0, ErrInvalidChunkSize},
		{"overlap equals chunk size", 100, 100, ErrInvalidOverlap},
		{"overlap exceeds chunk size", 100, 120, ErrInvalidOverlap},
		{"negative overlap", 100, -1, ErrInvalidOverlap},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTextChunker(tt.size, tt.overlap)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestChunkDocument(t *testing.T) {
	chunker, err := NewTextChunker(30, 5)
	if err != nil {
		t.Fatal(err)
	}

	doc := document.Document{
		ID: "doc-123",
		Text: "The quick brown fox jumps over the lazy dog. " +
			"This is another sentence for testing chunks.",
		Pages: []document.PageContent{
			{
				PageNumber: intPointer(1),
				Text: "The quick brown fox jumps over the lazy dog. " +
					"This is another sentence for testing chunks.",
			},
		},
	}

	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	for i, chunk := range chunks {
		if chunk.DocumentID != doc.ID {
			t.Errorf("unexpected document ID: %s", chunk.DocumentID)
		}

		if chunk.Index != i+1 {
			t.Errorf("expected index %d, got %d", i+1, chunk.Index)
		}

		if chunk.ID == "" {
			t.Error("chunk ID should not be empty")
		}

		if strings.TrimSpace(chunk.Text) == "" {
			t.Error("chunk text should not be empty")
		}

		if chunk.PageNumber == nil || *chunk.PageNumber != 1 {
			t.Error("expected page number 1")
		}

		if chunk.EndChar <= chunk.StartChar {
			t.Errorf("invalid character offsets: %+v", chunk)
		}
	}
}

func TestChunkDocumentEmptyText(t *testing.T) {
	chunker, _ := NewTextChunker(100, 20)

	chunks, err := chunker.ChunkDocument(document.Document{
		ID:   "doc-empty",
		Text: "   ",
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) != 0 {
		t.Fatalf("expected no chunks, got %d", len(chunks))
	}
}

func TestChunkDocumentFallsBackToDocumentText(t *testing.T) {
	chunker, _ := NewTextChunker(100, 10)

	doc := document.Document{
		ID:   "doc-fallback",
		Text: "This document has no page metadata.",
	}

	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %d", len(chunks))
	}

	if chunks[0].Text != doc.Text {
		t.Errorf("unexpected chunk text: %q", chunks[0].Text)
	}

	if chunks[0].PageNumber != nil {
		t.Error("expected no page number")
	}
}

func TestChunkDocumentPreservesPageNumbers(t *testing.T) {
	chunker, _ := NewTextChunker(100, 10)

	doc := document.Document{
		ID:   "doc-pages",
		Text: "Content from page one. Content from page two.",
		Pages: []document.PageContent{
			{PageNumber: intPointer(1), Text: "Content from page one."},
			{PageNumber: intPointer(2), Text: "Content from page two."},
		},
	}

	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected two chunks, got %d", len(chunks))
	}

	if *chunks[0].PageNumber != 1 || *chunks[1].PageNumber != 2 {
		t.Error("page numbers were not preserved")
	}
}

func TestChunkDocumentHandlesUnicode(t *testing.T) {
	chunker, _ := NewTextChunker(20, 5)

	doc := document.Document{
		ID:   "doc-unicode",
		Text: strings.Repeat("नमस्ते दुनिया ", 5),
	}

	chunks, err := chunker.ChunkDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	for _, chunk := range chunks {
		if !strings.Contains(chunk.Text, "नमस्ते") &&
			!strings.Contains(chunk.Text, "दुनिया") {
			t.Errorf("unexpected Unicode chunk: %q", chunk.Text)
		}
	}
}

func intPointer(value int) *int {
	return &value
}
