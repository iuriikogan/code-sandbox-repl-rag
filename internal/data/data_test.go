package data

import (
	"strings"
	"testing"
)

func TestGenerateUltraMassiveContext(t *testing.T) {
	content := GenerateUltraMassiveContext()
	size := len(content)
	t.Logf("Generated context size: %d bytes (~%.2f MB)", size, float64(size)/(1024*1024))
	
	if size < 15*1024*1024 {
		t.Errorf("Expected at least 15MB of context, got %d bytes", size)
	}

	if !strings.Contains(content, "Patient A (Male, 28)") {
		t.Error("Medical scenario marker missing")
	}
	if !strings.Contains(content, "Service Alpha code snippet") {
		t.Error("Engineering scenario marker missing")
	}
}
