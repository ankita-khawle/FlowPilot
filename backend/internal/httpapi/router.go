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

	"flowpilot/backend/internal/document"
)

const (
	maxUploadSize  = 10 << 20 // 10 MB
	maxRequestSize = maxUploadSize + (1 << 20)
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type DocumentExtractor interface {
	Extract(
		ctx context.Context,
		fileName string,
		content []byte,
	) (document.Document, error)
}

// Retain these types for the existing summarization API.
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

type Citation struct {
	Source string `json:"source"`
	Page   *int   `json:"page,omitempty"`
	Quote  string `json:"quote,omitempty"`
	Field  string `json:"field,omitempty"`
}

type ExtractedDocument struct {
	DocumentType string         `json:"document_type,omitempty"`
	Data         map[string]any `json:"data"`
}

type DocumentSummaryResponse struct {
	FileName      string            `json:"file_name"`
	Summary       string            `json:"summary"`
	ExtractedData ExtractedDocument `json:"extracted_data"`
	Citations     []Citation        `json:"citations"`
}

type DocumentProcessor interface {
	SummarizeDocument(
		ctx context.Context,
		fileName string,
		contentType string,
		fileContent []byte,
		tone SummaryTone,
		length SummaryLength,
	) (DocumentSummaryResponse, error)
}

// NewRouter configures the default document parser.
func NewRouter() http.Handler {
	return NewRouterWithServices(
		document.NewService(document.NewParser()),
		nil,
	)
}

func NewRouterWithServices(
	extractor DocumentExtractor,
	processor DocumentProcessor,
) http.Handler {
	return NewRouterWithRAGPipeline(extractor, processor, nil)
}

func NewRouterWithRAGPipeline(
	extractor DocumentExtractor,
	processor DocumentProcessor,
	pipeline RAGPipeline,
) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)

	mux.HandleFunc(
		"POST /api/documents",
		documentUpload(extractor),
	)

	mux.HandleFunc(
		"POST /api/documents/summarize",
		documentSummarize(processor),
	)

	// Register the RAG routes.
	handler := &Handler{
		ragPipeline: pipeline,
	}

	mux.HandleFunc(
		"POST /api/rag/documents",
		handler.ragUpload,
	)

	mux.HandleFunc(
		"POST /api/rag/documents/{documentID}/ask",
		handler.ragAsk,
	)

	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Service:   "flowpilot-api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func documentUpload(
	extractor DocumentExtractor,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if extractor == nil {
			writeError(
				w,
				http.StatusServiceUnavailable,
				"Document extraction service is not configured",
			)
			return
		}

		fileName, content, status, err := readUploadedDocument(w, r)
		if err != nil {
			writeError(w, status, err.Error())
			return
		}

		result, err := extractor.Extract(
			r.Context(),
			fileName,
			content,
		)

		if err != nil {
			switch {
			case errors.Is(err, document.ErrUnsupportedFormat):
				writeError(w, http.StatusBadRequest, "Unsupported file type")

			case errors.Is(err, document.ErrInvalidDocument):
				writeError(w, http.StatusBadRequest, "Invalid or corrupted document")

			case errors.Is(err, document.ErrEmptyDocument):
				writeError(w, http.StatusUnprocessableEntity, "No extractable text found in document")

			case errors.Is(err, document.ErrDocumentTooLarge):
				writeError(w, http.StatusRequestEntityTooLarge, "Document exceeds extraction limits")

			case errors.Is(err, context.Canceled):
				writeError(w, http.StatusRequestTimeout, "Document processing was cancelled")

			default:
				writeError(w, http.StatusInternalServerError, "Failed to extract document content")
			}

			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func documentSummarize(
	processor DocumentProcessor,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if processor == nil {
			writeError(
				w,
				http.StatusServiceUnavailable,
				"Document processing service is not configured",
			)
			return
		}

		fileName, content, status, err := readUploadedDocument(w, r)
		if err != nil {
			writeError(w, status, err.Error())
			return
		}

		tone := SummaryTone(r.FormValue("tone"))
		length := SummaryLength(r.FormValue("length"))

		if !isValidTone(tone) {
			writeError(w, http.StatusBadRequest, "Invalid summary tone")
			return
		}

		if !isValidLength(length) {
			writeError(w, http.StatusBadRequest, "Invalid summary length")
			return
		}

		result, err := processor.SummarizeDocument(
			r.Context(),
			fileName,
			detectContentType(filepath.Ext(fileName)),
			content,
			tone,
			length,
		)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to process document")
			return
		}

		result.FileName = fileName

		writeJSON(w, http.StatusOK, result)
	}
}

func readUploadedDocument(
	w http.ResponseWriter,
	r *http.Request,
) (string, []byte, int, error) {

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			return "", nil, http.StatusRequestEntityTooLarge,
				errors.New("request exceeds the maximum allowed size")
		}

		return "", nil, http.StatusBadRequest,
			errors.New("invalid multipart form")
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return "", nil, http.StatusBadRequest,
			errors.New("upload a document using the 'file' field")
	}
	defer file.Close()

	fileName := strings.TrimSpace(
		strings.ReplaceAll(header.Filename, "\\", "/"),
	)

	fileName = filepath.Base(fileName)

	if fileName == "" || fileName == "." {
		return "", nil, http.StatusBadRequest,
			errors.New("invalid filename")
	}

	extension := strings.ToLower(filepath.Ext(fileName))

	if !isSupportedDocument(extension) {
		return "", nil, http.StatusBadRequest,
			errors.New("unsupported file type. Upload PDF, DOCX, or TXT")
	}

	if header.Size > maxUploadSize {
		return "", nil, http.StatusRequestEntityTooLarge,
			errors.New("file exceeds the maximum allowed size of 10 MB")
	}

	content, err := io.ReadAll(
		io.LimitReader(file, maxUploadSize+1),
	)
	if err != nil {
		return "", nil, http.StatusBadRequest,
			errors.New("could not read uploaded document")
	}

	if len(content) == 0 {
		return "", nil, http.StatusBadRequest,
			errors.New("uploaded document is empty")
	}

	if len(content) > maxUploadSize {
		return "", nil, http.StatusRequestEntityTooLarge,
			errors.New("file exceeds the maximum allowed size of 10 MB")
	}

	return fileName, content, http.StatusOK, nil
}

func isSupportedDocument(extension string) bool {
	switch strings.ToLower(extension) {
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

func detectContentType(extension string) string {
	switch strings.ToLower(extension) {
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
