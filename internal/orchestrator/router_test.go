package orchestrator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

func setupRouterTest(t *testing.T) (*ai.Client, *Router) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT is not set. Skipping live router test.")
	}

	ctx := context.Background()
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	client, err := ai.NewClient(ctx, projectID, location)
	if err != nil {
		t.Fatalf("Failed to initialize AI client: %v", err)
	}

	runner := python.NewRunner()
	router := NewRouter(client, runner)
	return client, router
}

func TestClassifyQuery(t *testing.T) {
	client, router := setupRouterTest(t)
	defer client.Close()

	ctx := context.Background()

	// Test 1: A simple query
	simpleQuery := "What is the maximum vacation days allowed for employees per year?"
	complexity, reason, err := router.ClassifyQuery(ctx, simpleQuery)
	if err != nil {
		t.Fatalf("Failed to classify simple query: %v", err)
	}

	t.Logf("Simple Query classified as: %s (Reason: %s)", complexity, reason)
	if complexity != "SIMPLE" {
		t.Errorf("Expected SIMPLE classification, got %s", complexity)
	}

	// Test 2: A complex multi-hop relational query
	complexQuery := "Trace the genetic link between Patient A, Patient B and Patient C, and explain why Patient C was admitted to the ER."
	complexity, reason, err = router.ClassifyQuery(ctx, complexQuery)
	if err != nil {
		t.Fatalf("Failed to classify complex query: %v", err)
	}

	t.Logf("Complex Query classified as: %s (Reason: %s)", complexity, reason)
	if complexity != "COMPLEX" {
		t.Errorf("Expected COMPLEX classification, got %s", complexity)
	}
}

func TestRouteAndExecute_SimpleSmall(t *testing.T) {
	client, router := setupRouterTest(t)
	defer client.Close()

	ctx := context.Background()

	// Write a very small context file containing simple information
	simpleContext := "Employee vacation limit is 15 days per year.\nHR Policy version 2.1.\n"
	tmpFile, err := os.CreateTemp("", "simple-context-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(simpleContext); err != nil {
		t.Fatalf("Failed to write context: %v", err)
	}
	tmpFile.Close()

	query := "How many vacation days can employees take per year?"
	output, err := router.RouteAndExecute(ctx, tmpFile.Name(), query)
	if err != nil {
		t.Fatalf("RouteAndExecute failed: %v", err)
	}

	t.Logf("Direct Path Output:\n%s", output)
	if !strings.Contains(strings.ToLower(output), "15") {
		t.Errorf("Expected output to contain '15', got: %s", output)
	}
}
