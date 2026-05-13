package orchestrator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

func TestSyntheticCascadingFailureRAG(t *testing.T) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT not set, skipping live synthetic log RAG test")
	}

	ctx := context.Background()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		t.Fatalf("Failed to initialize AI client: %v", err)
	}
	defer client.Close()

	// 1. Generate 50 synthetic JSONL cascading failure logs
	t.Log("Generating synthetic cascading failure dataset via Vertex AI Gemini...")
	logs, err := data.GenerateCascadingFailureLogs(ctx, client, 50)
	if err != nil {
		t.Fatalf("Failed to generate synthetic logs: %v", err)
	}

	if len(logs) == 0 {
		t.Fatalf("Generated 0 logs")
	}

	tmpFilePath, cleanup, err := data.CreateJSONLContextFile(logs)
	if err != nil {
		t.Fatalf("Failed to create JSONL file: %v", err)
	}
	defer cleanup()

	// 2. Run Orchestrator Router against the synthetic JSONL dataset
	t.Log("Initializing Orchestrator against synthetic JSONL dataset...")
	runner := python.NewRunner()
	orch := New(client, runner)

	query := "Trace the root cause of the cascading failure. Identify connection pool exhaustion, Redis eviction spikes, or database connection lockups during the peak incident window."
	
	// Start orchestrator
	output, err := orch.Start(ctx, tmpFilePath, query)
	if err != nil {
		t.Fatalf("Orchestrator failed during synthetic evaluation: %v", err)
	}

	t.Logf("Orchestrator Final Synthesis Output:\n%s", output)

	outputLower := strings.ToLower(output)
	if !strings.Contains(outputLower, "connection") && !strings.Contains(outputLower, "timeout") && !strings.Contains(outputLower, "redis") && !strings.Contains(outputLower, "sql") {
		t.Errorf("Failed to synthesize key cascading failure insights. Output: %s", output)
	} else {
		t.Log("Successfully extracted and synthesized cascading failure root cause from synthetic LLM logs!")
	}
}
