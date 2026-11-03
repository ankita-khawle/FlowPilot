package config

import (
	"os"
	"testing"
)

func TestLoadUsesDefaultAddress(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	got := Load()
	if got.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want %q", got.HTTPAddr, ":8080")
	}
}

func TestLoadUsesEnvironmentAddress(t *testing.T) {
	const want = "127.0.0.1:9090"
	if err := os.Setenv("HTTP_ADDR", want); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("HTTP_ADDR") })

	got := Load()
	if got.HTTPAddr != want {
		t.Fatalf("HTTPAddr = %q, want %q", got.HTTPAddr, want)
	}
}
