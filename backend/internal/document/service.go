package document

import (
	"context"
	"crypto/rand"
	"fmt"
	"path"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	parser Parser
}

func NewService(parser Parser) *Service {
	return &Service{
		parser: parser,
	}
}

func (s *Service) Extract(
	ctx context.Context,
	fileName string,
	content []byte,
) (Document, error) {

	if s.parser == nil {
		return Document{}, fmt.Errorf("document parser is not configured")
	}

	if err := ctx.Err(); err != nil {
		return Document{}, err
	}

	fileName = strings.TrimSpace(
		strings.ReplaceAll(fileName, "\\", "/"),
	)

	fileName = path.Base(fileName)

	if fileName == "" || fileName == "." {
		return Document{}, fmt.Errorf("invalid filename")
	}

	extension := strings.ToLower(path.Ext(fileName))

	parsed, err := s.parser.Parse(ctx, fileName, content)
	if err != nil {
		return Document{}, fmt.Errorf("extract document: %w", err)
	}

	id, err := generateDocumentID()
	if err != nil {
		return Document{}, fmt.Errorf("generate document ID: %w", err)
	}

	return Document{
		ID:             id,
		FileName:       fileName,
		FileType:       strings.TrimPrefix(extension, "."),
		ContentType:    detectContentType(extension),
		Status:         "completed",
		Text:           parsed.Text,
		Pages:          parsed.Pages,
		WordCount:      len(strings.Fields(parsed.Text)),
		CharacterCount: utf8.RuneCountInString(parsed.Text),
		ExtractedAt:    time.Now().UTC(),
	}, nil
}

func generateDocumentID() (string, error) {
	var id [16]byte

	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}

	// UUID v4 version and variant bits.
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		id[0:4],
		id[4:6],
		id[6:8],
		id[8:10],
		id[10:16],
	), nil
}

func detectContentType(extension string) string {
	switch extension {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
