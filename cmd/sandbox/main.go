package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/orchestrator"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ui"
)

func main() {
	datasetPath := flag.String("dataset", "", "Absolute or relative path to an external dataset/log file (if empty, generates synthetic 45MB context)")
	customPrompt := flag.String("prompt", "", "Custom query instruction for the Orchestrator (if empty, uses default multi-scenario prompt)")
	flag.Parse()

	ctx := context.Background()

	// Initialize slog
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		slog.Error("Please set GOOGLE_CLOUD_PROJECT environment variable")
		os.Exit(1)
	}

	// Use us-central1 as it is the only region supporting Code Execution currently, but allow override
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	// Initialize the Vertex AI client wrapper
	slog.Info("Initializing GenAI client...", "projectID", projectID, "location", location)
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		slog.Error("Failed to initialize GenAI client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	var contextFilePath string
	var cleanup func()

	if *datasetPath != "" {
		if _, err := os.Stat(*datasetPath); err != nil {
			slog.Error("Provided dataset file does not exist", "path", *datasetPath, "error", err)
			os.Exit(1)
		}
		contextFilePath = *datasetPath
		cleanup = func() {} // No cleanup needed for user-provided files
		slog.Info("Using external dataset", "path", contextFilePath)
	} else {
		// Generate default synthetic dataset
		spinner := ui.NewSpinner("Generating 45MB ultra-massive context dataset...")
		spinner.Start()
		contextContent := data.GenerateUltraMassiveContext(1200000)
		contextFilePath, cleanup, err = data.CreateContextFile(contextContent)
		if err != nil {
			spinner.Stop("")
			slog.Error("Failed to create context file", "error", err)
			os.Exit(1)
		}
		spinner.Stop("✓ Dataset generated.")
	}
	defer cleanup()

	// Initialize the Python script runner locally with IPC
	slog.Info("Initializing Local IPC runner...")
	runner := python.NewRunner()

	// Start the tiered routing process
	slog.Info("Starting Tiered Routing...")
	router := orchestrator.NewRouter(client, runner)

	prompt := `Begin your task. Write a Python script to search 'context.txt' and extract the answers for TWO complex scenarios:
1. Medical: Trace the genetic link between Patient A, B, and C, and explain the acute ER admission of Patient C.
2. Engineering: Identify the root cause of the OOM kills in Service Omega, including the triggering service and proxy issue.`

	if *customPrompt != "" {
		prompt = *customPrompt
		slog.Info("Using custom query prompt", "prompt", prompt)
	}

	if _, err := router.RouteAndExecute(ctx, contextFilePath, prompt); err != nil {
		slog.Error("Router finished with error", "error", err)
		os.Exit(1)
	}
	slog.Info("Router finished successfully.")
}
