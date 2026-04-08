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

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ui"
)

type SandboxRunner struct {
	projectID  string
	location   string
	httpClient *http.Client
	engineID   string
}

func NewSandboxRunner(ctx context.Context, projectID, location string) (*SandboxRunner, error) {
	if location == "global" {
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

func (r *SandboxRunner) ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error) {
	if r.engineID == "" {
		spinner := ui.NewSpinner("Looking up deployed Reasoning Engine...")
		spinner.Start()
		id, err := r.getReasoningEngine(ctx)
		spinner.Stop("")
		if err != nil {
			return "", err
		}
		r.engineID = id
	}

	spinner := ui.NewSpinner("Provisioning secure Vertex AI Sandbox container...")
	spinner.Start()
	sandboxID, err := r.createSandbox(ctx)
	if err != nil {
		spinner.Stop("")
		return "", fmt.Errorf("failed to create sandbox: %w", err)
	}
	spinner.Stop("✓ Secure Sandbox provisioned.")
	defer r.deleteSandbox(ctx, sandboxID)

	spinner = ui.NewSpinner("Reading 45MB context file...")
	spinner.Start()
	contextContent, err := os.ReadFile(contextFileName)
	if err != nil {
		spinner.Stop("")
		return "", fmt.Errorf("failed to read context file: %w", err)
	}
	spinner.Stop("✓ Context read.")

	injectedCode := fmt.Sprintf(`import os
os.environ["PROJECT_ID"] = "%s"
os.environ["LOCATION"] = "%s"
import vertexai
vertexai.init(project="%s", location="%s")
%s`, r.projectID, r.location, r.projectID, r.location, code)

	spinner = ui.NewSpinner("Executing Python (hybrid search over 45MB data) in Sandbox via :executeCode...")
	spinner.Start()
	out, err := r.executeCode(ctx, sandboxID, injectedCode, contextContent)
	spinner.Stop("")
	if err != nil {
		return out, err
	}
	slog.Info("✓ Cloud Sandbox execution complete.")
	return out, nil
}

func (r *SandboxRunner) getReasoningEngine(ctx context.Context) (string, error) {
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
			return re.Name, nil
		}
	}
	return "", fmt.Errorf("Reasoning Engine '%s' not found. Please run setup_sandbox.sh first", displayName)
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

func (r *SandboxRunner) deleteSandbox(ctx context.Context, sandboxName string) {
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1beta1/%s", r.location, sandboxName)
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
