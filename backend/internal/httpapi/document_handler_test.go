package httpapi

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocumentUploadTXT(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "sample.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = part.Write([]byte("FlowPilot AI document extraction"))
	if err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		body,
	)

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	response := httptest.NewRecorder()

	NewRouter().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if !strings.Contains(response.Body.String(), "FlowPilot AI document extraction") {
		t.Fatal("expected extracted document text in response")
	}

	if !strings.Contains(response.Body.String(), `"status":"completed"`) {
		t.Fatal("expected completed processing status")
	}
}

func TestDocumentUploadUnsupportedFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "sample.xlsx")
	if err != nil {
		t.Fatal(err)
	}

	_, err = part.Write([]byte("test"))
	if err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		body,
	)

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	response := httptest.NewRecorder()

	NewRouter().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestDocumentUploadMissingFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("description", "test"); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		body,
	)

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	response := httptest.NewRecorder()

	NewRouter().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
