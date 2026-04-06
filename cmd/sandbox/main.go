package main

import (
	"context"
	"log/slog"
	"os"

	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/orchestrator"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		slog.Error("Please set GOOGLE_CLOUD_PROJECT environment variable")
		os.Exit(1)
	}

	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1" // Fallback to the default Vertex AI region
	}

	// Initialize the Vertex AI client wrapper
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		slog.Error("Failed to initialize GenAI client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Create a temporary file for the massive context data
	contextContent := data.GenerateMassiveContext()
	contextFilePath, cleanup, err := data.CreateContextFile(contextContent)
	if err != nil {
		slog.Error("Failed to create context file", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// Initialize the Python script runner
	runner := python.NewRunner()

	// Start the orchestration loop
	orch := orchestrator.New(client, runner)
	if err := orch.Start(ctx, contextFilePath); err != nil {
		slog.Error("Orchestrator finished with error", "error", err)
		os.Exit(1)
	}
}
