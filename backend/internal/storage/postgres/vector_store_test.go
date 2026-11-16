package postgres

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"flowpilot/backend/internal/rag"
)

func TestNewVectorStore(t *testing.T) {
	_, err := NewVectorStore(nil, 1536)
	if err == nil {
		t.Fatal("expected error for nil connection pool")
	}

	_, err = NewVectorStore(nil, 0)
	if err == nil {
		t.Fatal("expected error for invalid dimensions")
	}
}

func TestFormatVector(t *testing.T) {
	store := &VectorStore{dimensions: 3}

	got, err := store.formatVector([]float32{0.1, 0.2, 0.3})
	if err != nil {
		t.Fatal(err)
	}

	expected := "[0.1,0.2,0.3]"
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestFormatVectorInvalidDimensions(t *testing.T) {
	store := &VectorStore{dimensions: 3}

	_, err := store.formatVector([]float32{0.1, 0.2})

	if !errors.Is(err, rag.ErrInvalidVector) {
		t.Fatalf("expected ErrInvalidVector, got %v", err)
	}
}

func TestFormatVectorRejectsNonFiniteValues(t *testing.T) {
	store := &VectorStore{dimensions: 2}

	tests := []struct {
		name   string
		vector []float32
	}{
		{"NaN", []float32{float32(math.NaN()), 0.2}},
		{"positive infinity", []float32{float32(math.Inf(1)), 0.2}},
		{"negative infinity", []float32{float32(math.Inf(-1)), 0.2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.formatVector(tt.vector)

			if !errors.Is(err, rag.ErrInvalidVector) {
				t.Fatalf("expected ErrInvalidVector, got %v", err)
			}
		})
	}
}

func TestFormatVectorPreservesValues(t *testing.T) {
	store := &VectorStore{dimensions: 3}
	vector := []float32{0.25, -0.5, 1.0}

	encoded, err := store.formatVector(vector)
	if err != nil {
		t.Fatal(err)
	}

	if encoded != "[0.25,-0.5,1]" {
		t.Fatalf("unexpected encoding: %s", encoded)
	}

	if !reflect.DeepEqual(vector, []float32{0.25, -0.5, 1.0}) {
		t.Fatal("original vector was unexpectedly modified")
	}
}
