package orchestrator

import (
	"testing"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

func TestNewOrchestrator(t *testing.T) {
	// Simple test to ensure orchestrator struct creates correctly
	runner := python.NewRunner()
	orch := New(nil, runner)
	if orch == nil {
		t.Fatal("Expected new orchestrator, got nil")
	}
	if orch.runner == nil {
		t.Error("Expected runner to be assigned")
	}
}
