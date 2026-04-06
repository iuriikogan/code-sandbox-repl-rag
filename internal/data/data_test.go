package data

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateMassiveContext(t *testing.T) {
	ctx := GenerateMassiveContext()
	if !strings.Contains(ctx, "John Doe") {
		t.Errorf("Expected context to contain 'John Doe'")
	}
	if len(ctx) < 1000 {
		t.Errorf("Expected context to be massive, got length %d", len(ctx))
	}
}

func TestCreateContextFile(t *testing.T) {
	content := "test content data"
	filePath, cleanup, err := CreateContextFile(content)
	if err != nil {
		t.Fatalf("Failed to create context file: %v", err)
	}
	defer cleanup()

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", filePath)
	}

	// Verify content
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(bytes) != content {
		t.Errorf("Expected %q, got %q", content, string(bytes))
	}

	// Verify cleanup
	cleanup()
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Expected file to be deleted, but it still exists at %s", filePath)
	}
}
