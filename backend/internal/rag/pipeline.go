package rag

import (
	"context"
	"fmt"

	"flowpilot/backend/internal/ai"
	"flowpilot/backend/internal/document"
)

const embeddingBatchSize = 64

type DocumentExtractor interface {
	Extract(
		ctx context.Context,
		fileName string,
		content []byte,
	) (document.Document, error)
}

type IngestResult struct {
	DocumentID     string `json:"document_id"`
	FileName       string `json:"file_name"`
	FileType       string `json:"file_type"`
	Status         string `json:"status"`
	WordCount      int    `json:"word_count"`
	CharacterCount int    `json:"character_count"`
	ChunkCount     int    `json:"chunk_count"`
}

type Pipeline struct {
	extractor  DocumentExtractor
	chunker    Chunker
	embeddings ai.EmbeddingProvider
	store      VectorStore
	answerer   Answerer
}

func NewPipeline(
	extractor DocumentExtractor,
	chunker Chunker,
	embeddings ai.EmbeddingProvider,
	store VectorStore,
	answerer Answerer,
) (*Pipeline, error) {
	if extractor == nil {
		return nil, fmt.Errorf("document extractor is required")
	}
	if chunker == nil {
		return nil, fmt.Errorf("document chunker is required")
	}
	if embeddings == nil {
		return nil, fmt.Errorf("embedding provider is required")
	}
	if store == nil {
		return nil, fmt.Errorf("vector store is required")
	}
	if answerer == nil {
		return nil, fmt.Errorf("answerer is required")
	}

	return &Pipeline{
		extractor:  extractor,
		chunker:    chunker,
		embeddings: embeddings,
		store:      store,
		answerer:   answerer,
	}, nil
}

func (p *Pipeline) Ingest(
	ctx context.Context,
	fileName string,
	content []byte,
) (IngestResult, error) {
	if p == nil {
		return IngestResult{}, fmt.Errorf("RAG pipeline is nil")
	}

	doc, err := p.extractor.Extract(ctx, fileName, content)
	if err != nil {
		return IngestResult{}, fmt.Errorf("extract document: %w", err)
	}

	chunks, err := p.chunker.ChunkDocument(doc)
	if err != nil {
		return IngestResult{}, fmt.Errorf("chunk document: %w", err)
	}

	// An empty document can be ingested, but it will not have
	// searchable chunks.
	if len(chunks) == 0 {
		return buildIngestResult(doc, 0), nil
	}

	// Embed in batches to avoid sending too many inputs in one request.
	// Keep all vectors in memory and persist the chunks in one
	// VectorStore operation.
	vectors := make([][]float32, 0, len(chunks))

	for start := 0; start < len(chunks); start += embeddingBatchSize {
		if err := ctx.Err(); err != nil {
			return IngestResult{}, fmt.Errorf("embedding cancelled: %w", err)
		}

		end := start + embeddingBatchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := make([]string, 0, end-start)
		for _, chunk := range chunks[start:end] {
			batch = append(batch, chunk.Text)
		}

		batchVectors, err := p.embeddings.Embed(ctx, batch)
		if err != nil {
			return IngestResult{}, fmt.Errorf("embed document chunks: %w", err)
		}

		if len(batchVectors) != len(batch) {
			return IngestResult{}, fmt.Errorf(
				"embedding provider returned %d vectors for %d chunks",
				len(batchVectors),
				len(batch),
			)
		}

		vectors = append(vectors, batchVectors...)
	}

	if err := p.store.UpsertChunks(ctx, chunks, vectors); err != nil {
		return IngestResult{}, fmt.Errorf("store document chunks: %w", err)
	}

	return buildIngestResult(doc, len(chunks)), nil
}

func (p *Pipeline) Ask(
	ctx context.Context,
	documentID string,
	question string,
	limit int,
) (AnswerResponse, error) {
	if p == nil || p.answerer == nil {
		return AnswerResponse{}, fmt.Errorf("RAG answerer is unavailable")
	}

	return p.answerer.Answer(ctx, documentID, question, limit)
}

func buildIngestResult(
	doc document.Document,
	chunkCount int,
) IngestResult {
	return IngestResult{
		DocumentID:     doc.ID,
		FileName:       doc.FileName,
		FileType:       doc.FileType,
		Status:         doc.Status,
		WordCount:      doc.WordCount,
		CharacterCount: doc.CharacterCount,
		ChunkCount:     chunkCount,
	}
}
