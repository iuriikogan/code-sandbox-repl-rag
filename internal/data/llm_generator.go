package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"google.golang.org/genai"
)

// SyntheticLogEntry defines the structured format for our generated JSONL logs.
type SyntheticLogEntry struct {
	Timestamp   string         `json:"timestamp"`
	ServiceName string         `json:"service_name"`
	Severity    string         `json:"severity"`
	Message     string         `json:"message"`
	Payload     map[string]any `json:"payload,omitempty"`
}

// GenerateCascadingFailureLogs utilizes Vertex AI Gemini to generate structured JSONL logs representing a cascading failure.
func GenerateCascadingFailureLogs(ctx context.Context, client *ai.Client, numLogs int) ([]string, error) {
	startTime := time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC)
	peakStart := time.Date(2026, 5, 13, 8, 15, 0, 0, time.UTC)
	peakEnd := time.Date(2026, 5, 13, 8, 30, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC)
	totalDuration := endTime.Sub(startTime)

	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"timestamp":    {Type: genai.TypeString, Description: "ISO-8601 timestamp matching the requested time."},
			"service_name": {Type: genai.TypeString, Description: "Service name: 'gke-api-server', 'cloud-sql-proxy', or 'redis-cache'."},
			"severity":     {Type: genai.TypeString, Enum: []string{"INFO", "WARNING", "ERROR", "FATAL"}},
			"message":      {Type: genai.TypeString, Description: "Detailed log message reflecting the operational state."},
			"payload": {
				Type:        genai.TypeObject,
				Description: "Contextual metadata (e.g., pod_name, client_ip, query, query_latency_ms, memory_used_mb).",
			},
		},
		Required: []string{"timestamp", "service_name", "severity", "message"},
	}

	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(float32(0.4)),
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	results := make([]string, numLogs)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 15) // Concurrency semaphore cap at 15 to respect quotas
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	slog.Info("Starting LLM synthetic cascading failure log generation...", "targetLogs", numLogs)

	for i := 0; i < numLogs; i++ {
		wg.Add(1)
		timeOffset := time.Duration(r.Int63n(int64(totalDuration)))
		currentTime := startTime.Add(timeOffset)
		isPeak := currentTime.After(peakStart) && currentTime.Before(peakEnd)
		serviceChoice := r.Intn(3) // 0: GKE, 1: SQL, 2: Redis

		go func(index int, logTime time.Time, peak bool, serviceIdx int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()

			timeStr := logTime.Format(time.RFC3339Nano)
			var prompt string

			switch serviceIdx {
			case 0: // GKE
				if peak {
					prompt = fmt.Sprintf("Generate a GKE application pod log at %s during a peak incident. The pod is experiencing Redis connection pool exhaustion and database connection timeouts. Output severity ERROR or WARNING.", timeStr)
				} else {
					prompt = fmt.Sprintf("Generate a standard GKE application pod log at %s. HTTP 200 requests completed normally. Output severity INFO.", timeStr)
				}
			case 1: // Cloud SQL
				if peak {
					prompt = fmt.Sprintf("Generate a Cloud SQL Proxy log at %s during a cascading failure. Queries are timing out (>5000ms) and connection slots are fully reserved. Output severity ERROR or FATAL.", timeStr)
				} else {
					prompt = fmt.Sprintf("Generate a normal Cloud SQL query log at %s. Queries completing under 20ms. Output severity INFO.", timeStr)
				}
			case 2: // Redis
				if peak {
					prompt = fmt.Sprintf("Generate a Cloud MemoryStore Redis log at %s during memory pressure. Memory usage is >98%%, high key eviction rates, and OOM warnings. Output severity WARNING or ERROR.", timeStr)
				} else {
					prompt = fmt.Sprintf("Generate a standard Redis cache log at %s. Normal keyspace hits and stable memory. Output severity INFO.", timeStr)
				}
			}

			content := &genai.Content{
				Role:  "user",
				Parts: []*genai.Part{genai.NewPartFromText(prompt)},
			}

			// Use gemini-3.1-flash-preview for structured generation
			resp, err := client.GenAIClient.Models.GenerateContent(ctx, "gemini-3.1-flash-preview", []*genai.Content{content}, config)
			if err == nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
				part := resp.Candidates[0].Content.Parts[0]
				if part.Text != "" {
					// Validate JSON
					var entry SyntheticLogEntry
					if json.Unmarshal([]byte(part.Text), &entry) == nil {
						// Re-marshal into compact single-line JSONL
						compactBytes, _ := json.Marshal(entry)
						results[index] = string(compactBytes)
						return
					}
				}
			}

			// Fallback static log entry if LLM quota/rate limit is hit
			fb := SyntheticLogEntry{
				Timestamp:   timeStr,
				ServiceName: "gke-api-server",
				Severity:    "INFO",
				Message:     "Fallback static log entry due to generation limit.",
			}
			if peak {
				fb.Severity = "ERROR"
				fb.Message = "Cascading failure connection timeout (fallback)."
			}
			fbBytes, _ := json.Marshal(fb)
			results[index] = string(fbBytes)

		}(i, currentTime, isPeak, serviceChoice)
	}

	wg.Wait()

	// Filter out empty results if any
	var validLogs []string
	for _, log := range results {
		if log != "" {
			validLogs = append(validLogs, log)
		}
	}

	slog.Info("Successfully generated synthetic JSONL cascading failure dataset.", "totalLogs", len(validLogs))
	return validLogs, nil
}

// CreateJSONLContextFile writes the JSONL lines to a temporary file.
func CreateJSONLContextFile(logs []string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "cascading-*.jsonl")
	if err != nil {
		return "", nil, err
	}

	for _, log := range logs {
		if _, err := tmpFile.WriteString(log + "\n"); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", nil, err
		}
	}

	tmpFile.Close()
	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}
