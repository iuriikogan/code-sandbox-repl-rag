package ai

import (
	"context"
	"testing"
)

// Since cloud.google.com/go/vertexai/genai requires actual GCP credentials,
// running a full end-to-end test in CI or a generic environment will fail 
// unless GOOGLE_CLOUD_PROJECT is set and Application Default Credentials exist.
// This test provides basic sanity checks for the constants and types.
func TestClientConstants(t *testing.T) {
	if OrchestratorModelName == "" {
		t.Error("OrchestratorModelName should not be empty")
	}
	if WorkerModelName == "" {
		t.Error("WorkerModelName should not be empty")
	}
	if EmbeddingModelName == "" {
		t.Error("EmbeddingModelName should not be empty")
	}
}

func TestNewClient_MissingProject(t *testing.T) {
    // If we pass an empty project ID, it should attempt to create a client 
    // but might fail due to credential resolution depending on the env. 
    // We just ensure it doesn't panic.
    ctx := context.Background()
    client, _ := NewClient(ctx, "", "us-central1")
    if client != nil {
        client.Close()
    }
}
