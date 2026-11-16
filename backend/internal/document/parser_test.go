package document

import (
	"archive/zip"
	"context"
	"strings"
	"testing"
)

func TestParseTXT(t *testing.T) {
	parser := NewParser()

	result, err := parser.Parse(
		context.Background(),
		"sample.txt",
		[]byte("FlowPilot document extraction test"),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Text != "FlowPilot document extraction test" {
		t.Fatalf("unexpected text: %s", result.Text)
	}

	if len(result.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(result.Pages))
	}

	if result.Pages[0].PageNumber == nil ||
		*result.Pages[0].PageNumber != 1 {
		t.Fatal("expected page number 1")
	}
}

func TestParseUnsupportedFormat(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse(
		context.Background(),
		"sample.xlsx",
		[]byte("content"),
	)

	if err != ErrUnsupportedFormat {
		t.Fatalf("expected unsupported format, got %v", err)
	}
}

func TestParseEmptyTXT(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse(
		context.Background(),
		"empty.txt",
		[]byte("   "),
	)

	if err != ErrEmptyDocument {
		t.Fatalf("expected empty document error, got %v", err)
	}
}

func TestParseInvalidTXT(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse(
		context.Background(),
		"invalid.txt",
		[]byte{0xff, 0xfe, 0xfd},
	)

	if err != ErrInvalidDocument {
		t.Fatalf("expected invalid document error, got %v", err)
	}
}

func TestParseInvalidPDF(t *testing.T) {
	parser := NewParser()

	_, err := parser.Parse(
		context.Background(),
		"invalid.pdf",
		[]byte("this is not a PDF"),
	)

	if err != ErrInvalidDocument {
		t.Fatalf("expected invalid PDF error, got %v", err)
	}
}

func TestParseDOCX(t *testing.T) {
	parser := NewParser()

	content := createTestDOCX(t)

	result, err := parser.Parse(
		context.Background(),
		"sample.docx",
		content,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Text, "Employee Agreement") {
		t.Fatalf("expected document title in extracted text: %s", result.Text)
	}

	if !strings.Contains(result.Text, "Annual salary") {
		t.Fatalf("expected paragraph in extracted text: %s", result.Text)
	}

	if len(result.Pages) != 1 {
		t.Fatalf("expected one logical content block, got %d", len(result.Pages))
	}

	if result.Pages[0].PageNumber != nil {
		t.Fatal("DOCX should not have an invented page number")
	}
}

func createTestDOCX(t *testing.T) []byte {
	t.Helper()

	var buffer strings.Builder

	writer := zip.NewWriter(&stringWriter{builder: &buffer})

	file, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
		<w:p><w:r><w:t>Employee Agreement</w:t></w:r></w:p>
		<w:p><w:r><w:t>Annual salary: 2500000 INR</w:t></w:r></w:p>
	</w:body>
</w:document>`

	if _, err := file.Write([]byte(xmlContent)); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	return []byte(buffer.String())
}

// stringWriter adapts strings.Builder to io.Writer.
type stringWriter struct {
	builder *strings.Builder
}

func (w *stringWriter) Write(p []byte) (int, error) {
	return w.builder.Write(p)
}
