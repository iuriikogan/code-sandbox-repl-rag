package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

// TestAccuracy_RAG benchmarks the retrieval accuracy of the Agentic RAG approach.
// It inserts a specific, unique "needle" into a large "haystack" and verifies if the 
// orchestrator can find it.
func TestAccuracy_RAG(t *testing.T) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT not set, skipping integration accuracy test")
	}

	ctx := context.Background()
	location := "us-central1"

	// 1. Initialize Client and Runner
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		t.Fatalf("Failed to initialize GenAI client: %v", err)
	}
	
	// Use LocalRunner for testability in standard environments
	runner := python.NewRunner()
	orch := New(client, runner)

	// 2. Create Haystack with a Needle
	// We'll use 100 repeats of standard data and insert one unique finance event.
	baseHaystack := data.GenerateContext(100)
	needle := "FINANCE EVENT: On July 14th, the company acquired 'AI-Solutions Inc' for $450M in cash."
	
	// Inject the needle in the middle
	parts := strings.SplitAfterN(baseHaystack, "\n", 300)
	fullContext := ""
	if len(parts) > 300 {
		fullContext = strings.Join(parts[:300], "") + needle + "\n" + strings.Join(parts[300:], "")
	} else {
		fullContext = baseHaystack + needle + "\n"
	}

	contextFilePath, cleanup, err := data.CreateContextFile(fullContext)
	if err != nil {
		t.Fatalf("Failed to create context file: %v", err)
	}
	defer cleanup()

	// 3. Run Orchestrator
	// Capture stdout to verify the final synthesis
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// We'll need a way to stop the orchestrator or just let it finish.
	// Since Start() is a blocking call that finishes when the final synthesis is done,
	// we can just wait for it.
	err = orch.Start(ctx, contextFilePath)
	
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Orchestrator failed: %v", err)
	}

	// 4. Validate Results
	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	fmt.Println("--- DEBUG OUTPUT ---")
	fmt.Println(output)

	if !strings.Contains(strings.ToLower(output), "ai-solutions") || !strings.Contains(output, "$450m") {
		t.Errorf("Accuracy Failure: The orchestrator failed to find or synthesize the finance needle.\nOutput: %s", output)
	} else {
		t.Log("Accuracy Success: Needle found and synthesized correctly.")
	}
}
