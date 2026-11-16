package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"flowpilot/backend/internal/rag"
)

type VectorStore struct {
	pool       *pgxpool.Pool
	dimensions int
}

func NewVectorStore(
	pool *pgxpool.Pool,
	dimensions int,
) (*VectorStore, error) {
	if pool == nil {
		return nil, errors.New("database connection pool is required")
	}

	if dimensions <= 0 {
		return nil, errors.New("embedding dimensions must be positive")
	}

	return &VectorStore{
		pool:       pool,
		dimensions: dimensions,
	}, nil
}

func (s *VectorStore) UpsertChunks(
	ctx context.Context,
	chunks []rag.Chunk,
	vectors [][]float32,
) error {
	if len(chunks) != len(vectors) {
		return rag.ErrVectorCountMismatch
	}

	if len(chunks) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin vector storage transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, chunk := range chunks {
		if chunk.ID == "" || chunk.DocumentID == "" {
			return fmt.Errorf("chunk %d has missing identifiers", i)
		}

		if strings.TrimSpace(chunk.Text) == "" {
			return fmt.Errorf("chunk %d has empty content", i)
		}

		vector, err := s.formatVector(vectors[i])
		if err != nil {
			return fmt.Errorf("validate vector for chunk %s: %w", chunk.ID, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO document_chunks (
				chunk_id,
				document_id,
				chunk_index,
				content,
				page_number,
				start_char,
				end_char,
				embedding
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8::vector)
			ON CONFLICT (chunk_id)
			DO UPDATE SET
				document_id = EXCLUDED.document_id,
				chunk_index = EXCLUDED.chunk_index,
				content = EXCLUDED.content,
				page_number = EXCLUDED.page_number,
				start_char = EXCLUDED.start_char,
				end_char = EXCLUDED.end_char,
				embedding = EXCLUDED.embedding
		`,
			chunk.ID,
			chunk.DocumentID,
			chunk.Index,
			chunk.Text,
			chunk.PageNumber,
			chunk.StartChar,
			chunk.EndChar,
			vector,
		)
		if err != nil {
			return fmt.Errorf("store chunk %s: %w", chunk.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit vector storage transaction: %w", err)
	}

	return nil
}

func (s *VectorStore) Search(
	ctx context.Context,
	queryVector []float32,
	documentID string,
	limit int,
) ([]rag.ChunkMatch, error) {
	if strings.TrimSpace(documentID) == "" {
		return nil, rag.ErrEmptyDocumentID
	}

	if limit <= 0 {
		return nil, rag.ErrInvalidSearchLimit
	}

	vector, err := s.formatVector(queryVector)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			chunk_id,
			document_id,
			chunk_index,
			content,
			page_number,
			start_char,
			end_char,
			1 - (embedding <=> $1::vector) AS similarity
		FROM document_chunks
		WHERE document_id = $2
		ORDER BY embedding <=> $1::vector
		LIMIT $3
	`, vector, documentID, limit)
	if err != nil {
		return nil, fmt.Errorf("search document vectors: %w", err)
	}
	defer rows.Close()

	var matches []rag.ChunkMatch

	for rows.Next() {
		var match rag.ChunkMatch
		var pageNumber *int

		if err := rows.Scan(
			&match.Chunk.ID,
			&match.Chunk.DocumentID,
			&match.Chunk.Index,
			&match.Chunk.Text,
			&pageNumber,
			&match.Chunk.StartChar,
			&match.Chunk.EndChar,
			&match.Similarity,
		); err != nil {
			return nil, fmt.Errorf("scan vector search result: %w", err)
		}

		match.Chunk.PageNumber = pageNumber
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vector search results: %w", err)
	}

	return matches, nil
}

func (s *VectorStore) formatVector(
	vector []float32,
) (string, error) {
	if len(vector) != s.dimensions {
		return "", fmt.Errorf(
			"%w: expected %d dimensions, got %d",
			rag.ErrInvalidVector,
			s.dimensions,
			len(vector),
		)
	}

	values := make([]string, len(vector))

	for i, value := range vector {
		if math.IsNaN(float64(value)) ||
			math.IsInf(float64(value), 0) {
			return "", fmt.Errorf(
				"%w: non-finite value at index %d",
				rag.ErrInvalidVector,
				i,
			)
		}

		values[i] = strconv.FormatFloat(float64(value), 'g', -1, 32)
	}

	return "[" + strings.Join(values, ",") + "]", nil
}
