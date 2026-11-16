package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"flowpilot/backend/internal/document"
	"flowpilot/backend/internal/rag"
)

type Handler struct {
	ragPipeline RAGPipeline
}

const (
	defaultRetrievalLimit = 5
	maxRetrievalLimit     = 20
	maxQuestionLength     = 4000
)

type RAGPipeline interface {
	Ingest(
		ctx context.Context,
		fileName string,
		content []byte,
	) (rag.IngestResult, error)

	Ask(
		ctx context.Context,
		documentID string,
		question string,
		limit int,
	) (rag.AnswerResponse, error)
}

type askRequest struct {
	Question string `json:"question"`
	Limit    int    `json:"limit,omitempty"`
}

func (h *Handler) ragUpload(
	w http.ResponseWriter,
	r *http.Request,
) {
	if h.ragPipeline == nil {
		writeError(w, http.StatusServiceUnavailable, "RAG pipeline is unavailable")
		return
	}

	fileName, content, status, err := readRAGUpload(w, r)
	if err != nil {
		writeError(w, status, err.Error())
		return
	}

	result, err := h.ragPipeline.Ingest(r.Context(), fileName, content)
	if err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, document.ErrUnsupportedFormat):
			status = http.StatusBadRequest
		case errors.Is(err, document.ErrInvalidDocument),
			errors.Is(err, document.ErrEmptyDocument):
			status = http.StatusUnprocessableEntity
		case errors.Is(err, document.ErrDocumentTooLarge):
			status = http.StatusRequestEntityTooLarge
		}

		writeError(w, status, "Unable to ingest document")
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *Handler) ragAsk(
	w http.ResponseWriter,
	r *http.Request,
) {
	if h.ragPipeline == nil {
		writeError(w, http.StatusServiceUnavailable, "RAG pipeline is unavailable")
		return
	}

	documentID := strings.TrimSpace(r.PathValue("documentID"))
	if documentID == "" {
		writeError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request askRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Reject additional JSON values after the first request object.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "Request must contain one JSON object")
		return
	}

	request.Question = strings.TrimSpace(request.Question)

	if request.Question == "" {
		writeError(w, http.StatusBadRequest, "Question is required")
		return
	}

	if len([]rune(request.Question)) > maxQuestionLength {
		writeError(w, http.StatusBadRequest, "Question is too long")
		return
	}

	if request.Limit == 0 {
		request.Limit = defaultRetrievalLimit
	}

	if request.Limit < 1 || request.Limit > maxRetrievalLimit {
		writeError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf("Limit must be between 1 and %d", maxRetrievalLimit),
		)
		return
	}

	result, err := h.ragPipeline.Ask(
		r.Context(),
		documentID,
		request.Question,
		request.Limit,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to generate answer")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func readRAGUpload(
	w http.ResponseWriter,
	r *http.Request,
) (string, []byte, int, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return "", nil, http.StatusRequestEntityTooLarge, errors.New("upload exceeds maximum request size")
		}

		return "", nil, http.StatusBadRequest, errors.New("invalid multipart upload")
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return "", nil, http.StatusBadRequest, errors.New("file is required")
	}
	defer file.Close()

	fileName := filepath.Base(strings.ReplaceAll(header.Filename, "\\", "/"))

	if fileName == "." || fileName == "" {
		return "", nil, http.StatusBadRequest, errors.New("invalid file name")
	}

	if !isSupportedDocument(filepath.Ext(fileName)) {
		return "", nil, http.StatusBadRequest, errors.New("unsupported document format")
	}

	content, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil {
		return "", nil, http.StatusBadRequest, errors.New("unable to read uploaded file")
	}

	if len(content) == 0 {
		return "", nil, http.StatusBadRequest, errors.New("uploaded file is empty")
	}

	if len(content) > maxUploadSize {
		return "", nil, http.StatusRequestEntityTooLarge, errors.New("file exceeds maximum upload size")
	}

	return fileName, content, http.StatusOK, nil
}
