package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"google.golang.org/genai"
)

// Router handles tiered model routing logic based on query complexity and context size.
type Router struct {
	client *ai.Client
	runner python.Runner
}

// NewRouter creates a new Router instance.
func NewRouter(client *ai.Client, runner python.Runner) *Router {
	return &Router{
		client: client,
		runner: runner,
	}
}

// ClassificationResult matches the structured JSON schema for complexity classification.
type ClassificationResult struct {
	Complexity string `json:"complexity"` // "SIMPLE" or "COMPLEX"
	Reason     string `json:"reason"`
}

// ClassifyQuery classifies a prompt into SIMPLE or COMPLEX using Tier 1 (Flash-Lite).
func (r *Router) ClassifyQuery(ctx context.Context, query string) (string, string, error) {
	prompt := fmt.Sprintf(`Analyze the following user query and determine its complexity.
A query is SIMPLE if it asks for a single fact, a basic lookup, or generic information that does not require tracing relationships between multiple different logs/records.
A query is COMPLEX if it requires multi-hop reasoning, tracing timelines or relations across multiple files/services, analyzing memory leaks, diagnostic logs, or medical history across patients.

QUERY:
%s`, query)

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(float32(0.0)),
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"complexity": {
					Type:        genai.TypeString,
					Enum:        []string{"SIMPLE", "COMPLEX"},
					Description: "The complexity level of the query.",
				},
				"reason": {
					Type:        genai.TypeString,
					Description: "Reasoning behind the classification.",
				},
			},
			Required: []string{"complexity", "reason"},
		},
	}

	slog.Info("Classifying query complexity...", "model", ai.Tier1ModelName)
	resp, err := r.client.GenAIClient.Models.GenerateContent(ctx, ai.Tier1ModelName, []*genai.Content{content}, config)
	if err != nil {
		return "", "", fmt.Errorf("failed to classify query: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", "", fmt.Errorf("empty response from classification model")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text == "" {
		return "", "", fmt.Errorf("classification response is not text")
	}

	var res ClassificationResult
	if err := json.Unmarshal([]byte(part.Text), &res); err != nil {
		return "", "", fmt.Errorf("failed to parse classification json: %w, raw: %s", err, part.Text)
	}

	return res.Complexity, res.Reason, nil
}

// RouteAndExecute chooses the correct tiered execution path and returns the final answer.
func (r *Router) RouteAndExecute(ctx context.Context, contextFilePath string, query string) (string, error) {
	// 1. Estimate Context Size
	info, err := os.Stat(contextFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat context file: %w", err)
	}
	contextSize := info.Size()
	slog.Info("Estimated context size", "bytes", contextSize, "mb", float64(contextSize)/(1024*1024))

	// 2. Classify Complexity
	complexity, reason, err := r.ClassifyQuery(ctx, query)
	if err != nil {
		slog.Warn("Query classification failed, defaulting to COMPLEX path", "error", err)
		complexity = "COMPLEX"
		reason = "fallback"
	}
	slog.Info("Query classification complete", "complexity", complexity, "reason", reason)

	// Thresholds
	const (
		SmallThreshold  = 100 * 1024 // 100 KB (~20,000 tokens)
		MediumThreshold = 1 * 1024 * 1024 // 1 MB (~200,000 tokens)
	)

	// 3. Route to Tier
	if contextSize < SmallThreshold {
		if complexity == "SIMPLE" {
			slog.Info("Routing to DIRECT TIER 1 PATH (Gemini 3.1 Flash-Lite)", "size", contextSize)
			return r.executeDirectPath(ctx, ai.Tier1ModelName, contextFilePath, query)
		} else {
			slog.Info("Routing to DIRECT TIER 3 PATH (Gemini 3.1 Pro) for small complex context", "size", contextSize)
			return r.executeDirectPath(ctx, ai.Tier3ModelName, contextFilePath, query)
		}
	}

	if complexity == "SIMPLE" && contextSize < MediumThreshold {
		slog.Info("Routing to DIRECT TIER 2 PATH (Gemini 3.1 Flash)", "size", contextSize)
		return r.executeDirectPath(ctx, ai.Tier2ModelName, contextFilePath, query)
	}

	// Fall through or explicit complex massive path -> Agentic Swarm RAG
	slog.Info("Routing to AGENTIC SWARM PATH (Orchestrator + Python + Swarm Workers)", "size", contextSize, "complexity", complexity)
	orch := New(r.client, r.runner)
	return orch.Start(ctx, contextFilePath, query)
}

func (r *Router) executeDirectPath(ctx context.Context, modelName string, contextFilePath string, query string) (string, error) {
	contentBytes, err := os.ReadFile(contextFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read context file: %w", err)
	}
	contextText := string(contentBytes)

	prompt := fmt.Sprintf(`You are an expert customer engineering assistant. Answer the query below strictly using the provided context. Do not hallucinate.

CONTEXT:
%s

QUERY:
%s`, contextText, query)

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.1)),
	}

	resp, err := r.client.GenAIClient.Models.GenerateContent(ctx, modelName, []*genai.Content{content}, config)
	if err != nil {
		return "", fmt.Errorf("direct generation failed with model %s: %w", modelName, err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from model %s", modelName)
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text == "" {
		return "", fmt.Errorf("response from model %s is not text", modelName)
	}

	return part.Text, nil
}
