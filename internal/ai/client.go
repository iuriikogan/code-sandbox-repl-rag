package ai

import (
	"context"

	"fmt"

	"log/slog"

	"os"

	"sync"

	"google.golang.org/genai"
)

const (
	WorkerModelName         = "gemini-2.5-flash"
	FinalSynthesisModelName = "gemini-2.5-pro"
	EmbeddingModelName      = "text-embedding-004" // Default for Vertex
)

// Client wraps the standard GenAI client.
type Client struct {
	GenAIClient           *genai.Client
	EmbedModel            string
	OrchestratorModelName string
}

// NewClient initializes a new AI client.
func NewClient(ctx context.Context, projectID, location string) (*Client, error) {
	var client *genai.Client
	var err error
	embedModel := EmbeddingModelName
	orchestratorModel := "gemini-2.5-flash"

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" {
		slog.Info("Using Gemini Developer API (AI Studio) due to GEMINI_API_KEY")
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: apiKey,
		})
		embedModel = "text-embedding-004" // Use standard text-embedding-004 for AI Studio
		orchestratorModel = "gemini-3.1-flash-lite-preview"
	} else {
		slog.Info("Using Vertex AI Backend")
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			Backend:  genai.BackendVertexAI,
			Project:  projectID,
			Location: location,
		})
	}

	if err != nil {
		return nil, err
	}
	return &Client{
		GenAIClient:           client,
		EmbedModel:            embedModel,
		OrchestratorModelName: orchestratorModel,
	}, nil
}

// Close closes the underlying client.

func (c *Client) Close() error {

	return nil

}

// HandleBatchCall runs a swarm of worker agents concurrently for a slice of data chunks.

func (c *Client) HandleBatchCall(ctx context.Context, instruction string, contextChunks []string) []string {

	results := make([]string, len(contextChunks))

	var wg sync.WaitGroup

	// Limit to 10 concurrent requests to respect rate limits

	sem := make(chan struct{}, 10)

	for i, chunk := range contextChunks {

		wg.Add(1)

		go func(index int, data string) {

			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}: // Acquire token
			}

			results[index] = c.HandleCall(ctx, instruction, data)

			<-sem // Release token

		}(i, chunk)

	}

	wg.Wait()

	return results

}

// HandleBatchEmbed generates vector embeddings efficiently in bulk API calls (chunked to respect limits).

func (c *Client) HandleBatchEmbed(ctx context.Context, texts []string) [][]float32 {

	var allVectors [][]float32

	const batchSize = 100

	for i := 0; i < len(texts); i += batchSize {

		end := i + batchSize

		if end > len(texts) {

			end = len(texts)

		}

		var contents []*genai.Content

		for _, t := range texts[i:end] {

			contents = append(contents, &genai.Content{

				Parts: []*genai.Part{genai.NewPartFromText(t)},
			})

		}

		res, err := c.GenAIClient.Models.EmbedContent(ctx, c.EmbedModel, contents, nil)

		if err != nil {

			slog.Error("Failed to batch embed content", "error", err, "batch_start", i)

			// Pad with nils to maintain index alignment if a batch fails

			for j := i; j < end; j++ {

				allVectors = append(allVectors, nil)

			}

			continue

		}

		for _, emb := range res.Embeddings {

			allVectors = append(allVectors, emb.Values)

		}

	}

	return allVectors

}

// HandleCall runs a worker agent for a specific data chunk.
func (c *Client) HandleCall(ctx context.Context, instruction, contextChunk string) string {
	slog.Info("Spinning up Flash Sub-Agent", "chunk_size", len(contextChunk))

	prompt := fmt.Sprintf(`You are a specialized worker agent. Follow the instructions strictly.
	
INSTRUCTION: %s

DATA CHUNK:
%s`, instruction, contextChunk)

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	resp, err := c.GenAIClient.Models.GenerateContent(ctx, WorkerModelName, []*genai.Content{content}, nil)
	if err != nil {
		return fmt.Sprintf("Sub-agent failed: %v", err)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if part.Text != "" {
			return part.Text
		}
	}
	return "No response from sub-agent"
}

// HandleEmbed generates a vector embedding for a given text chunk.
func (c *Client) HandleEmbed(ctx context.Context, text string) []float32 {
	content := &genai.Content{
		Parts: []*genai.Part{genai.NewPartFromText(text)},
	}

	res, err := c.GenAIClient.Models.EmbedContent(ctx, c.EmbedModel, []*genai.Content{content}, nil)
	if err != nil {
		slog.Error("Failed to embed content", "error", err)
		return nil
	}

	if len(res.Embeddings) > 0 && len(res.Embeddings[0].Values) > 0 {
		return res.Embeddings[0].Values
	}
	return nil
}
