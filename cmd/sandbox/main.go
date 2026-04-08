package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/orchestrator"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ui"
)

func main() {
	ctx := context.Background()

	// Initialize slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		slog.Error("Please set GOOGLE_CLOUD_PROJECT environment variable")
		os.Exit(1)
	}

	// Use us-central1 as it is the only region supporting Code Execution currently
	const location = "us-central1"

	// Initialize the Vertex AI client wrapper
	slog.Info("Initializing GenAI client...", "projectID", projectID, "location", location)
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		slog.Error("Failed to initialize GenAI client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Create a temporary file for the massive context data
	spinner := ui.NewSpinner("Generating 45MB ultra-massive context dataset...")
	spinner.Start()
	contextContent := data.GenerateUltraMassiveContext()
	contextFilePath, cleanup, err := data.CreateContextFile(contextContent)
	if err != nil {
		spinner.Stop("")
		slog.Error("Failed to create context file", "error", err)
		os.Exit(1)
	}
	spinner.Stop("✓ Dataset generated.")
	defer cleanup()

	// Initialize the Python script runner (Vertex AI Agent Engine Sandbox)
	slog.Info("Initializing Cloud Sandbox runner (Reasoning Engine)...")
	runner, err := python.NewSandboxRunner(ctx, projectID, location)
	if err != nil {
		slog.Error("Failed to initialize Sandbox runner", "error", err)
		os.Exit(1)
	}

	// Start the orchestration loop
	slog.Info("Starting Orchestrator loop...")
	orch := orchestrator.New(client, runner)
	prompt := `Begin your task. Write a Python script to search 'context.txt' and extract the answers for TWO complex scenarios:
1. Medical: Trace the genetic link between Patient A, B, and C, and explain the acute ER admission of Patient C.
2. Engineering: Identify the root cause of the OOM kills in Service Omega, including the triggering service and proxy issue.`

	if err := orch.Start(ctx, contextFilePath, prompt); err != nil {
		slog.Error("Orchestrator finished with error", "error", err)
		os.Exit(1)
	}

	slog.Info("Orchestrator finished successfully.")
}
