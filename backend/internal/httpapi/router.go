package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const maxUploadSize = 10 << 20 // 10 MB

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type SummaryTone string
type SummaryLength string

const (
	ToneProfessional SummaryTone = "professional"
	ToneCasual       SummaryTone = "casual"
	ToneTechnical    SummaryTone = "technical"
	ToneExecutive    SummaryTone = "executive"

	LengthShort    SummaryLength = "short"
	LengthMedium   SummaryLength = "medium"
	LengthDetailed SummaryLength = "detailed"
)

// Citation identifies the source of an extracted fact or summary statement.
// The document processor should populate page numbers and excerpts when available.
type Citation struct {
	Source  string `json:"source"`
	Page    *int   `json:"page,omitempty"`
	Quote   string `json:"quote,omitempty"`
	Field   string `json:"field,omitempty"`
}

type ExtractedDocument struct {
	DocumentType string         `json:"document_type,omitempty"`
	Data         map[string]any `json:"data"`
}

type DocumentSummaryResponse struct {
	FileName     string           `json:"file_name"`
	Summary      string           `json:"summary"`
	ExtractedData ExtractedDocument `json:"extracted_data"`
	Citations    []Citation       `json:"citations"`
}

type DocumentProcessor interface {
	// SummarizeDocument should extract text from the uploaded file,
	// call the LLM, and return structured data with source citations.
	SummarizeDocument(
		ctx context.Context,
		fileName string,
		contentType string,
		fileContent []byte,
		tone SummaryTone,
		length SummaryLength,
	) (DocumentSummaryResponse, error)
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewRouter builds the HTTP handler without a document processor.
// The document endpoint returns 503 until a processor is configured.
func NewRouter() http.Handler {
	return NewRouterWithDocumentProcessor(nil)
}

// NewRouterWithDocumentProcessor builds the HTTP handler with document processing.
func NewRouterWithDocumentProcessor(processor DocumentProcessor) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("POST /api/documents/summarize", documentSummarize(processor))

	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   "flowpilot-api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func documentSummarize(processor DocumentProcessor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if processor == nil {
			writeError(w, http.StatusServiceUnavailable,
				"Document processing service is not configured")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+(1<<20))

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeError(w, http.StatusRequestEntityTooLarge,
					"Upload exceeds the maximum allowed size of 10 MB")
				return
			}

			writeError(w, http.StatusBadRequest,
				"Invalid multipart form")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest,
				"Please upload a document using the 'file' field")
			return
		}
		defer file.Close()

		extension := strings.ToLower(filepath.Ext(header.Filename))
		if !isSupportedDocument(extension) {
			writeError(w, http.StatusBadRequest,
				"Unsupported file type. Upload a PDF, DOCX, or TXT file")
			return
		}

		content, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				"Could not read uploaded document")
			return
		}

		if len(content) == 0 {
			writeError(w, http.StatusBadRequest,
				"The uploaded document is empty")
			return
		}

		tone := SummaryTone(r.FormValue("tone"))
		length := SummaryLength(r.FormValue("length"))

		if !isValidTone(tone) {
			writeError(w, http.StatusBadRequest,
				"Invalid tone. Use professional, casual, technical, or executive")
			return
		}

		if !isValidLength(length) {
			writeError(w, http.StatusBadRequest,
				"Invalid length. Use short, medium, or detailed")
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(content)
		}

		result, err := processor.SummarizeDocument(
			r.Context(),
			header.Filename,
			contentType,
			content,
			tone,
			length,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError,
				"Failed to process document")
			return
		}

		result.FileName = header.Filename

		writeJSON(w, http.StatusOK, result)
	}
}

func isSupportedDocument(extension string) bool {
	switch extension {
	case ".pdf", ".docx", ".txt":
		return true
	default:
		return false
	}
}

func isValidTone(tone SummaryTone) bool {
	switch tone {
	case ToneProfessional, ToneCasual, ToneTechnical, ToneExecutive:
		return true
	default:
		return false
	}
}

func isValidLength(length SummaryLength) bool {
	switch length {
	case LengthShort, LengthMedium, LengthDetailed:
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Error: message,
	})
}