package python

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

// SandboxRunner executes Python code in a Vertex AI Agent Engine sandbox.
type SandboxRunner struct {
	projectID  string
	location   string
	httpClient *http.Client
	engineID   string
}

// NewSandboxRunner creates a new SandboxRunner.
func NewSandboxRunner(ctx context.Context, projectID, location string) (*SandboxRunner, error) {
	if location == "global" {
		// Agent Engine Code Execution is currently only in us-central1
		location = "us-central1"
	}

	opts := []option.ClientOption{
		option.WithScopes("https://www.googleapis.com/auth/cloud-platform"),
	}
	httpClient, _, err := htransport.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	return &SandboxRunner{
		projectID:  projectID,
		location:   location,
		httpClient: httpClient,
	}, nil
}

// ExecuteScript runs a Python script in a Vertex AI Sandbox.
// It ignores the IPCHandler for now as we transition to a cloud-native model.
func (r *SandboxRunner) ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error) {
	slog.Info("Executing Python in Vertex AI Sandbox...")

	// 1. Get or Create Reasoning Engine
	if r.engineID == "" {
		id, err := r.getOrCreateReasoningEngine(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to ensure reasoning engine: %w", err)
		}
		r.engineID = id
	}

	// 2. Create Sandbox (or reuse if we want to be persistent, but for now we create one)
	sandboxID, err := r.createSandbox(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create sandbox: %w", err)
	}
	defer r.deleteSandbox(ctx, sandboxID)

	// 3. Read context file to pass into the sandbox
	contextContent, err := os.ReadFile(contextFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read context file: %w", err)
	}

	// 4. Inject project and location into the code
	injectedCode := fmt.Sprintf(`import os
os.environ['PROJECT_ID'] = "%s"
os.environ['LOCATION'] = "%s"
import vertexai
vertexai.init(project="%s", location="%s")
%s`, r.projectID, r.location, r.projectID, r.location, code)

	// 5. Execute Code
	return r.executeCode(ctx, sandboxID, injectedCode, contextContent)
}

func (r *SandboxRunner) getOrCreateReasoningEngine(ctx context.Context) (string, error) {
	// For simplicity, we search for one named "rag-simulation-engine" or create it.
	displayName := "rag-simulation-engine"
	parent := fmt.Sprintf("projects/%s/locations/%s", r.projectID, r.location)
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1beta1/%s/reasoningEngines", r.location, parent)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var listResp struct {
		ReasoningEngines []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"reasoningEngines"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return "", err
	}

	for _, re := range listResp.ReasoningEngines {
		if re.DisplayName == displayName {
			// Extract ID from full name
			return re.Name, nil
		}
	}

	// Create new one
	slog.Info("Creating new Reasoning Engine...")
	createURL := url
	payload := map[string]any{
		"display_name": displayName,
		"spec": map[string]any{
			"package_spec": map[string]any{
				"python_version": "3.10",
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	req, _ = http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err = r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create reasoning engine: %s", string(body))
	}

	var op struct {
		Name     string `json:"name"`
		Response struct {
			Name string `json:"name"`
		} `json:"response"`
		Done bool `json:"done"`
	}
	// Note: In a real app, you should wait for the LRO.
	// But for this simulation, we'll assume it's fast or just error out.
	if err := json.NewDecoder(resp.Body).Decode(&op); err != nil {
		return "", err
	}
	
	// If it's an LRO, we should wait.
	if !op.Done {
		// Just wait a bit for simulation purposes, or poll.
		slog.Info("Waiting for Reasoning Engine creation...")
		return "", fmt.Errorf("Reasoning Engine creation started (LRO: %s), please try again in a few seconds", op.Name)
	}

	return op.Response.Name, nil
}

func (r *SandboxRunner) createSandbox(ctx context.Context) (string, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1beta1/%s/sandboxes", r.location, r.engineID)
	payload := map[string]any{
		"config": map[string]any{
			"ttl": "3600s",
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create sandbox: %s", string(body))
	}

	var sandbox struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sandbox); err != nil {
		return "", err
	}
	return sandbox.Name, nil
}

func (r *SandboxRunner) deleteSandbox(ctx context.Context, name string) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1beta1/%s", r.location, name)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	r.httpClient.Do(req)
}

func (r *SandboxRunner) executeCode(ctx context.Context, sandboxName, code string, contextContent []byte) (string, error) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1beta1/%s:executeCode", r.location, sandboxName)
	
	payload := map[string]any{
		"input_data": map[string]any{
			"code": code,
			"files": []map[string]any{
				{
					"name":    "context.txt",
					"content": base64.StdEncoding.EncodeToString(contextContent),
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to execute code: %s", string(body))
	}

	var res struct {
		Outputs []struct {
			Content  string `json:"content"`
			MimeType string `json:"mimeType"`
		} `json:"outputs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	var combinedOutput string
	for _, out := range res.Outputs {
		decoded, _ := base64.StdEncoding.DecodeString(out.Content)
		combinedOutput += string(decoded)
	}

	return combinedOutput, nil
}
