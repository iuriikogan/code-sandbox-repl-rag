package orchestrator

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"google.golang.org/genai"
)

const (
	ProInputCostPer1M    = 3.50
	ProOutputCostPer1M   = 10.50
	FlashInputCostPer1M  = 0.075
	FlashOutputCostPer1M = 0.30
	EmbeddingCostPer1M   = 0.02
)

func calculateCost(modelName string, inputTokens, outputTokens int32) float64 {
	var inputCost, outputCost float64
	switch modelName {
	case ai.FinalSynthesisModelName: // 2.5-pro
		inputCost = float64(inputTokens) / 1000000.0 * ProInputCostPer1M
		outputCost = float64(outputTokens) / 1000000.0 * ProOutputCostPer1M
	case ai.WorkerModelName: // 2.5-flash
		inputCost = float64(inputTokens) / 1000000.0 * FlashInputCostPer1M
		outputCost = float64(outputTokens) / 1000000.0 * FlashOutputCostPer1M
	case ai.EmbeddingModelName:
		inputCost = float64(inputTokens) / 1000000.0 * EmbeddingCostPer1M
	}
	return inputCost + outputCost
}

func evaluateAccuracy(output string) int {
	output = strings.ToLower(output)
	expectedKeywords := []string{
		"x-trace", "omega", "oom-kill", "envoy", "memory leak",
		"rule 44b", "alpha", "ff_archive_sync", "alice",
		"payload", "2mb", "cgroup", "istio-proxy", "cron-beta",
	}
	found := 0
	for _, kw := range expectedKeywords {
		if strings.Contains(output, kw) {
			found++
		}
	}
	return found
}

func TestIntegratedEvaluationEngineering(t *testing.T) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("No project ID")
	}

	ctx := context.Background()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		t.Fatalf("Client error: %v", err)
	}
	defer client.Close()

	dataset := data.GenerateUltraMassiveContext(150000)
	tmpFile, cleanup, err := data.CreateContextFile(dataset)
	if err != nil {
		t.Fatalf("File error: %v", err)
	}
	defer cleanup()

	prompt := "Extract a summary of the root cause of the Service Omega outage.\n\nDATA:\n" + dataset

	t.Log("--- RUNNING SINGLE ENDPOINT (Gemini Pro) ---")
	content := &genai.Content{Role: "user", Parts: []*genai.Part{genai.NewPartFromText(prompt)}}
	config := &genai.GenerateContentConfig{Temperature: genai.Ptr(float32(0.2))}
	startPro := time.Now()
	respPro, err := client.GenAIClient.Models.GenerateContent(ctx, ai.FinalSynthesisModelName, []*genai.Content{content}, config)
	if err != nil {
		t.Logf("Pure Pro Failed (quota/size expected): %v", err)
	}
	if err == nil && len(respPro.Candidates) > 0 {
		var input, output int32
		if respPro.UsageMetadata != nil {
			input = respPro.UsageMetadata.PromptTokenCount
			output = respPro.UsageMetadata.CandidatesTokenCount
		}
		cost := calculateCost(ai.FinalSynthesisModelName, input, output)
		score := evaluateAccuracy(respPro.Candidates[0].Content.Parts[0].Text)
		t.Logf("Pure Pro: Score=%d/14, Cost=$%.6f, Latency=%v", score, cost, time.Since(startPro))
	}

	t.Log("--- RUNNING REPL RAG (Sandbox) ---")
	ragPrompt := "Find the root cause of the Service Omega outage. Trace the chain of events to underlying compliance rules, feature flags, payload issues, and proxy memory leaks."
	runner := python.NewRunner()
	orch := New(client, runner)
	startRAG := time.Now()
	finalOutput, err := orch.Start(ctx, tmpFile, ragPrompt)
	if err != nil {
		t.Fatalf("Orchestrator failed: %v", err)
	}

	ragScore := evaluateAccuracy(finalOutput)
	t.Logf("RAG Finished: Score=%d/14, Latency=%v\nOutput:\n%s", ragScore, time.Since(startRAG), finalOutput)
}

func TestIntegratedEvaluationMedical(t *testing.T) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("No project ID")
	}

	ctx := context.Background()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		t.Fatalf("Client error: %v", err)
	}
	defer client.Close()

	dataset := data.GenerateUltraMassiveContext(150000)
	tmpFile, cleanup, err := data.CreateContextFile(dataset)
	if err != nil {
		t.Fatalf("File error: %v", err)
	}
	defer cleanup()

	t.Log("--- RUNNING REPL RAG (Sandbox) MEDICAL ---")
	ragPrompt := "Trace the genetic link between Patient A, B, and C, and explain the acute ER admission of Patient C."
	runner := python.NewRunner()
	orch := New(client, runner)
	startRAG := time.Now()
	finalOutput, err := orch.Start(ctx, tmpFile, ragPrompt)
	if err != nil {
		t.Fatalf("Orchestrator failed: %v", err)
	}

	outputLower := strings.ToLower(finalOutput)
	medicalKeywords := []string{"patient a", "patient b", "patient c", "photosensitivity", "dark urine", "peripheral neuropathy", "sulfonamide", "neurovisceral", "acute intermittent porphyria", "aip"}

	ragScore := 0
	for _, kw := range medicalKeywords {
		if strings.Contains(outputLower, kw) {
			ragScore++
		}
	}
	t.Logf("RAG Finished: Score=%d/%d, Latency=%v\nOutput:\n%s", ragScore, len(medicalKeywords), time.Since(startRAG), finalOutput)
}
