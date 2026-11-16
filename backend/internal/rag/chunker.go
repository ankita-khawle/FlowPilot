package rag

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"flowpilot/backend/internal/document"
)

var (
	ErrInvalidChunkSize = errors.New("chunk size must be greater than zero")
	ErrInvalidOverlap   = errors.New("overlap must be smaller than chunk size")
)

type Chunk struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	Index      int    `json:"index"`
	Text       string `json:"text"`
	PageNumber *int   `json:"page_number,omitempty"`
	StartChar  int    `json:"start_char"`
	EndChar    int    `json:"end_char"`
}

type Chunker interface {
	ChunkDocument(doc document.Document) ([]Chunk, error)
}

type TextChunker struct {
	maxChars     int
	overlapChars int
}

func NewTextChunker(maxChars, overlapChars int) (*TextChunker, error) {
	if maxChars <= 0 {
		return nil, ErrInvalidChunkSize
	}

	if overlapChars < 0 || overlapChars >= maxChars {
		return nil, ErrInvalidOverlap
	}

	return &TextChunker{
		maxChars:     maxChars,
		overlapChars: overlapChars,
	}, nil
}

func (c *TextChunker) ChunkDocument(
	doc document.Document,
) ([]Chunk, error) {
	if c == nil {
		return nil, errors.New("chunker is nil")
	}

	if strings.TrimSpace(doc.Text) == "" {
		return []Chunk{}, nil
	}

	pages := doc.Pages
	if len(pages) == 0 {
		pages = []document.PageContent{
			{Text: doc.Text},
		}
	}

	var chunks []Chunk
	chunkIndex := 0
	globalOffset := 0

	for _, page := range pages {
		pageText := page.Text
		if strings.TrimSpace(pageText) == "" {
			globalOffset += len([]rune(pageText))
			continue
		}

		pageChunks := c.chunkText(pageText)

		for _, item := range pageChunks {
			chunkIndex++

			pageNumber := page.PageNumber
			chunks = append(chunks, Chunk{
				ID:         fmt.Sprintf("%s-chunk-%04d", doc.ID, chunkIndex),
				DocumentID: doc.ID,
				Index:      chunkIndex,
				Text:       item.text,
				PageNumber: pageNumber,
				StartChar:  globalOffset + item.start,
				EndChar:    globalOffset + item.end,
			})
		}

		globalOffset += len([]rune(pageText))
	}

	return chunks, nil
}

type textSegment struct {
	text  string
	start int
	end   int
}

func (c *TextChunker) chunkText(text string) []textSegment {
	runes := []rune(text)
	total := len(runes)

	var chunks []textSegment

	start := 0

	for start < total {
		end := start + c.maxChars
		if end > total {
			end = total
		}

		// Prefer ending at a word boundary.
		if end < total {
			for end > start && !unicode.IsSpace(runes[end]) &&
				!unicode.IsSpace(runes[end-1]) {
				end--
			}

			// Avoid producing a tiny chunk when a single word
			// is longer than the configured chunk size.
			if end == start {
				end = start + c.maxChars
				if end > total {
					end = total
				}
			}
		}

		chunkText := strings.TrimSpace(string(runes[start:end]))

		if chunkText != "" {
			chunks = append(chunks, textSegment{
				text:  chunkText,
				start: start,
				end:   end,
			})
		}

		if end >= total {
			break
		}

		nextStart := end - c.overlapChars
		if nextStart <= start {
			nextStart = end
		}

		start = nextStart
	}

	return chunks
}
