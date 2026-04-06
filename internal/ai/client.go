package ai

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
)

const (
	OrchestratorModelName   = "gemini-1.5-flash"
	WorkerModelName         = "gemini-1.5-flash"
	FinalSynthesisModelName = "gemini-1.5-pro"
	EmbeddingModelName      = "text-embedding-004"
)

// Client wraps the standard GenAI client.
type Client struct {
	GenAIClient *genai.Client
}

// NewClient initializes a new AI client.
func NewClient(ctx context.Context, projectID, location string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  projectID,
		Location: location,
	})
	if err != nil {
		return nil, err
	}
	return &Client{GenAIClient: client}, nil
}

// Close closes the underlying client.
func (c *Client) Close() error {
	return nil
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

	res, err := c.GenAIClient.Models.EmbedContent(ctx, EmbeddingModelName, []*genai.Content{content}, nil)
	if err != nil {
		slog.Error("Failed to embed content", "error", err)
		return nil
	}

	if len(res.Embeddings) > 0 && len(res.Embeddings[0].Values) > 0 {
		return res.Embeddings[0].Values
	}
	return nil
}