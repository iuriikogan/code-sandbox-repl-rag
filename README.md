# Code Sandbox REPL RAG

## Project Overview
This project is a Go-based agentic Retrieval-Augmented Generation (RAG) simulation. It demonstrates how to build an Orchestrator agent using Google's Gemini models (`gemini-2.5-flash`) that can autonomously execute generated Python code to process massive unstructured context data (approx 45-50MB / 1.2M+ lines).

The application leverages **Vertex AI Reasoning Engines** for secure, isolated Python execution in the cloud. Because the dataset far exceeds reasonable LLM context limits, the system forces a **Two-Stage Hybrid RAG approach**:
1. **Lexical Pre-filtering:** The generated Python script aggressively filters the millions of lines down to a few hundred chunks using regex/keyword matching in memory.
2. **Semantic Search:** The script then generates embeddings only for the filtered chunks via the Vertex AI Python SDK, calculates Cosine Similarity, and returns the top matches to the Go Orchestrator for final synthesis using `gemini-2.5-pro`.

### Key Technologies
- **Language**: Go (`go 1.25.0`)
- **SDK**: Google GenAI SDK (`google.golang.org/genai`)
- **Models Used**:
  - `gemini-2.5-flash` (Orchestrator)
  - `gemini-2.5-flash-lite` (Sub-agent worker)
  - `gemini-2.5-pro` (Final Synthesis)
  - `text-embedding-004` (Semantic search / embeddings)
- **Execution Environment**: Vertex AI Reasoning Engines (Cloud Sandbox)

## Architecture Details
1. **Context Generation**: Go dynamically creates a 45MB ultra-massive dataset simulating deep engineering memory leaks and multi-generational medical diagnostic data.
2. **GCS Upload**: The unstructured dataset is uploaded to a Google Cloud Storage bucket (`rag-sandbox-obj-{project}-us-central1`).
3. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-2.5-flash`, passing it an `run_rag_agent_logic` tool.
4. **Execution**: The generated Python script is sent to the deployed Vertex AI Reasoning Engine via the `:query` API path attached to a custom executor. It pulls the context from GCS, runs the hybrid RAG logic locally within the container, and returns the top highly-relevant chunks.
5. **Synthesis**: The Orchestrator receives the parsed, high-value chunks and feeds them into `gemini-2.5-pro` for a final, polished reasoning output.

## Prerequisites
- Go 1.25+
- Python 3 & pip
- Google Cloud Project with Vertex AI API enabled
- Authenticated via Application Default Credentials (`gcloud auth application-default login`)

### Environment Variables
You must set the following environment variable before running the application:
- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID.

Note: The application uses the **`us-central1`** Vertex AI endpoint exclusively, as it is the target region for Agent Engine features.

## Setup & Running

### 1. Provision the Cloud Sandbox
Before running the Go application, you **must** provision the Reasoning Engine and GCS bucket on your Google Cloud Project. This takes about 3-5 minutes.
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
./setup_sandbox.sh
```

### 2. Run the RAG Simulation
To run the Go orchestrator:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
go run cmd/sandbox/main.go
```

## Project Structure
- `cmd/sandbox/main.go`: Main application entry point.
- `internal/ai/`: Wrappers and clients for Google GenAI interactions.
- `internal/data/`: Massive dataset generator (1.2M+ lines / 45MB).
- `internal/orchestrator/`: Primary agent orchestration and GenAI loop.
- `internal/python/`: Execution logic for interfacing with Vertex AI Reasoning Engines and local fallback runners.
- `internal/ui/`: Terminal spinners and visual feedback.
