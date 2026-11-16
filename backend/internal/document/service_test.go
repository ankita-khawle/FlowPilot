package document

import (
	"context"
	"testing"
)

func TestServiceExtract(t *testing.T) {
	service := NewService(NewParser())

	result, err := service.Extract(
		context.Background(),
		"employment.txt",
		[]byte("Employee: John Doe\nSalary: 2500000 INR"),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID == "" {
		t.Fatal("expected generated document ID")
	}

	if result.FileName != "employment.txt" {
		t.Fatalf("unexpected filename: %s", result.FileName)
	}

	if result.FileType != "txt" {
		t.Fatalf("unexpected file type: %s", result.FileType)
	}

	if result.ContentType != "text/plain" {
		t.Fatalf("unexpected content type: %s", result.ContentType)
	}

	if result.Status != "completed" {
		t.Fatalf("unexpected status: %s", result.Status)
	}

	if result.WordCount != 6 {
		t.Fatalf("expected 6 words, got %d", result.WordCount)
	}

	if result.CharacterCount == 0 {
		t.Fatal("expected character count")
	}

	if result.ExtractedAt.IsZero() {
		t.Fatal("expected extraction timestamp")
	}
}

func TestServiceRejectsInvalidFilename(t *testing.T) {
	service := NewService(NewParser())

	_, err := service.Extract(
		context.Background(),
		".",
		[]byte("sample"),
	)

	if err == nil {
		t.Fatal("expected invalid filename error")
	}
}

func TestServiceRejectsCancelledContext(t *testing.T) {
	service := NewService(NewParser())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.Extract(
		ctx,
		"sample.txt",
		[]byte("sample"),
	)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
