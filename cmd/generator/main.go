package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/data"
)

func main() {
	numLogs := flag.Int("logs", 100, "Number of cascading failure logs to generate")
	outPath := flag.String("out", "cascading_failure.jsonl", "Output file path for the JSONL dataset")
	flag.Parse()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		slog.Error("Please set GOOGLE_CLOUD_PROJECT environment variable")
		os.Exit(1)
	}

	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	slog.Info("Initializing GenAI client...", "projectID", projectID, "location", location)
	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		slog.Error("Failed to initialize GenAI client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	slog.Info("Generating synthetic cascading failure logs...", "count", *numLogs)
	logs, err := data.GenerateCascadingFailureLogs(ctx, client, *numLogs)
	if err != nil {
		slog.Error("Failed to generate logs", "error", err)
		os.Exit(1)
	}

	file, err := os.Create(*outPath)
	if err != nil {
		slog.Error("Failed to create output file", "path", *outPath, "error", err)
		os.Exit(1)
	}
	defer file.Close()

	for _, log := range logs {
		if _, err := file.WriteString(log + "\n"); err != nil {
			slog.Error("Failed to write log line", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Cascading failure dataset generated successfully!", "path", *outPath, "totalLogs", len(logs))
}
