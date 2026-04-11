package orchestrator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"google.golang.org/genai"
)

func setupClient(t *testing.T) *ai.Client {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT is not set. Skipping live accuracy test.")
	}

	ctx := context.Background()
	// Use us-central1 for consistency and tool support, but allow override
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		t.Fatalf("Failed to initialize ai client: %v", err)
	}
	return client
}

func checkAccuracy(t *testing.T, output string) {
	output = strings.ToLower(output)
	// Engineering Scenario keywords (expanded for deeper multi-hop reasoning)
	expectedKeywords := []string{
		"x-trace",
		"omega",
		"oom-kill",
		"envoy",
		"memory leak",
		"rule 44b",
		"alpha",
		"ff_archive_sync",
		"alice",
		"payload",
		"2mb",
		"cgroup",
		"istio-proxy",
		"cron-beta",
	}

	found := 0
	for _, kw := range expectedKeywords {
		if strings.Contains(output, kw) {
			found++
		}
	}

	// We have 14 possible keywords. A highly accurate model traversing the full chain
	// should find at least 10 of these concepts.
	if found < 10 {
		t.Errorf("Accuracy failed: Expected at least 10 keywords from %v, but only found %d in output:\n%s", expectedKeywords, found, output)
	} else {
		t.Logf("Accuracy passed: Found %d keywords in output.", found)
	}
}

// TestAccuracy_PureFlash tests the accuracy of gemini-2.5-flash with a massive context.
func TestAccuracy_PureFlash(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	ctx := context.Background()
	dataset := data.GenerateUltraMassiveContext(300000)
	prompt := "Extract a summary of the root cause of the Service Omega outage.\n\nDATA:\n" + dataset

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
	}

	t.Log("Sending to gemini-2.5-flash...")
	resp, err := client.GenAIClient.Models.GenerateContent(ctx, ai.OrchestratorModelName, []*genai.Content{content}, config)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		t.Fatalf("Empty response")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text != "" {
		checkAccuracy(t, part.Text)
	} else {
		t.Fatalf("Response is not text")
	}
}

// TestAccuracy_PurePro tests the accuracy of gemini-2.5-pro with a massive context.
func TestAccuracy_PurePro(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	ctx := context.Background()
	dataset := data.GenerateUltraMassiveContext(300000)
	prompt := "Extract a summary of the root cause of the Service Omega outage.\n\nDATA:\n" + dataset

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
	}

	t.Log("Sending to gemini-2.5-pro...")
	resp, err := client.GenAIClient.Models.GenerateContent(ctx, ai.FinalSynthesisModelName, []*genai.Content{content}, config)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		t.Fatalf("Empty response")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text != "" {
		checkAccuracy(t, part.Text)
	} else {
		t.Fatalf("Response is not text")
	}
}

// TestAccuracy_RAGOrchestrator tests the accuracy of our orchestrated RAG approach.
func TestAccuracy_RAGOrchestrator(t *testing.T) {
	client := setupClient(t)
	defer client.Close()

	dataset := data.GenerateUltraMassiveContext(300000)
	tmpFile, cleanup, err := data.CreateContextFile(dataset)
	if err != nil {
		t.Fatalf("Failed to create context file: %v", err)
	}
	defer cleanup()

	// Ensure python path is correct for the sandbox runner.
	// Assuming we are running inside the /internal/orchestrator directory,
	// we might need to set up the python runner to use the venv in project root.
	runner := python.NewRunner()
	orch := New(client, runner)
	prompt := `Begin your task. Write a Python script to search the file at CONTEXT_FILE using lexical search to extract the initial chunks containing keywords related to the Service Omega outage (e.g. "Omega", "OOM-kill", "cron-beta"). Then, iteratively use semantic search embeddings via IPC to trace the root cause back to the original feature flag, payload changes, and compliance rules that triggered the envoy proxy memory leak.`
	t.Log("Starting Orchestrator RAG approach...")
	finalOutput, err := orch.Start(context.Background(), tmpFile, prompt)
	if err != nil {
		t.Fatalf("Orchestrator failed: %v", err)
	}

	checkAccuracy(t, finalOutput)
}
