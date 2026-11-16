
# FlowPilot API

Go backend for FlowPilot, an AI-powered document intelligence platform.

## Current capabilities

- Health check endpoint
- PDF, DOCX and TXT document uploads
- File size and format validation
- Text extraction
- PDF page-level extraction
- Document metadata and processing statistics
- Context-aware extraction cancellation
- Unit and HTTP handler tests

## Tech stack

- Go 1.23
- PDF text extraction: ledongthuc/pdf
- DOCX parsing: Go archive/zip and encoding/xml
- TXT parsing: Go standard library

## Run locally

From the repository root:

```bash
cd backend
go mod tidy
go run ./cmd/api
```

API runs at http://localhost:8080.

## Health check

```bash
curl http://localhost:8080/api/health
```

## Upload a document

```bash
curl -X POST http://localhost:8080/api/documents \
  -F "file=@sample.pdf"
```

Supported formats:

- PDF
- DOCX
- TXT

Maximum file size: 10 MB.

## Test

```bash
go test ./...
go vet ./...
```

## Current API endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | /api/health | Health check |
| POST | /api/documents | Extract document text |
| POST | /api/documents/summarize | Existing summary endpoint |

## Upcoming capabilities

- OpenAI and Claude integrations
- Structured data extraction
- Document chunking
- Embeddings and vector search
- RAG-based summarization
- Document Q&A
- AI-powered workflows

Database and background job integrations will be introduced in subsequent milestones.
