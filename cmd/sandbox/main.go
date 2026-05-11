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

	// Use us-central1 as it is the only region supporting Code Execution currently
	const location = "us-central1"

	// Initialize the Vertex AI client wrapper
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		slog.Error("Failed to initialize GenAI client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Create a temporary file for the massive context data (Simulated SEC Filing)
	contextContent := data.GenerateSECContext()
	contextFilePath, cleanup, err := data.CreateContextFile(contextContent)
	if err != nil {
		slog.Error("Failed to create context file", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// Initialize the Python script runner (GKE Sandbox / gVisor)
	namespace := os.Getenv("K8S_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	image := os.Getenv("WORKER_IMAGE")
	if image == "" {
		image = "gcr.io/your-project/python-worker:latest" // TODO: Update with real image
	}

	runner, err := python.NewGKERunner(ctx, namespace, image)
	if err != nil {
		slog.Error("Failed to initialize GKE runner", "error", err)
		os.Exit(1)
	}

	// Start the orchestration loop
	orch := orchestrator.New(client, runner)
	if err := orch.Start(ctx, contextFilePath); err != nil {
		slog.Error("Orchestrator finished with error", "error", err)
		os.Exit(1)
	}
}
