package document

import "time"

// PageContent represents extracted text and its original location.
// PageNumber is nil when the source format does not preserve pagination,
// such as DOCX.
type PageContent struct {
	PageNumber *int   `json:"page_number,omitempty"`
	Text       string `json:"text"`
}

// Document represents the normalized result of document ingestion.
type Document struct {
	ID             string        `json:"id"`
	FileName       string        `json:"file_name"`
	FileType       string        `json:"file_type"`
	ContentType    string        `json:"content_type"`
	Status         string        `json:"status"`
	Text           string        `json:"text"`
	Pages          []PageContent `json:"pages"`
	WordCount      int           `json:"word_count"`
	CharacterCount int           `json:"character_count"`
	ExtractedAt    time.Time     `json:"extracted_at"`
}

// ParsedContent is the parser's output before document metadata is added.
type ParsedContent struct {
	Text  string        `json:"text"`
	Pages []PageContent `json:"pages"`
}
