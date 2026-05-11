# Code Sandbox REPL RAG (Gemini 3.1 & GKE)

## Project Overview
**Code Sandbox REPL RAG** is an enterprise-grade Agentic Retrieval-Augmented Generation (RAG) system. It orchestrates a multi-model workflow to process massive, novel datasets (e.g., SEC 10-K filings) securely and cost-effectively by leveraging **GKE Sandbox (gVisor)** for isolated code execution.

### Key Pillars
- **Secure, Isolated Execution**: Uses **GKE Sandbox (gVisor)** to run AI-generated Python code in a hardened kernel environment, preventing container escape and protecting the host.
- **Tiered Discovery Workflow**: A cost-optimized pipeline that uses **Gemini 3.1 Flash-Lite** for initial triage and regex-based filtering, followed by **Gemini 3.1 Flash** for semantic search, and **Gemini 3.1 Pro** for final synthesis.
- **Agentic RAG**: Replaces static data pipelines with dynamic Python scripts that handle chunking, embedding generation, and local similarity search tailored to the specific dataset.
- **Massive Scalability**: Designed for "mega-datasets" where traditional RAG fails, shifting heavy computation (filtering/clustering) into the sandbox to minimize expensive LLM tokens.

### Benefits
- **Uncompromised Security**: Sandbox execution ensures enterprise-grade security even if the LLM generates destructive or flawed commands.
- **Extreme Cost Efficiency**: The Tiered Discovery approach is **~38% cheaper** than Naive RAG and over **260x cheaper** than Direct Long-Context Pro synthesis for large datasets.
- **Superior Reasoning**: By distilling the dataset before final processing, Gemini 3.1 Pro can focus its reasoning on the most relevant high-value data.

### Key Technologies
- **Language**: Go 1.26+
- **SDK**: Google GenAI SDK (`google.golang.org/genai`)
- **Infrastructure**: GKE (Standard or Autopilot) with `gvisor` RuntimeClass.
- **Models**:
  - `gemini-3.1-flash-lite` (Triage & Sub-agents)
  - `gemini-3.1-flash` (Orchestrator)
  - `gemini-3.1-pro` (Final Synthesis)
  - `text-embedding-004` (Embeddings)

## Architecture Details
1. **Orchestration**: Go initializes **Gemini 3.1 Flash** with Tiered Discovery instructions.
2. **Sandbox Execution**: Go creates an ephemeral **GKE Job** with the `gvisor` runtime. The Python code is injected via a **ConfigMap**.
3. **Tiered Processing**: 
   - Python performs rapid regex/keyword triage.
   - High-value blocks are evaluated by **Flash-Lite** sub-agents.
   - Relevant chunks are embedded and filtered via local Cosine Similarity.
4. **Synthesis**: The distilled "Insight Manifest" is returned to Go, which invokes **Gemini 3.1 Pro** for the final executive summary.

## Prerequisites
- Go 1.26+
- GKE Cluster with `gvisor` enabled.
- Google Cloud Project with Vertex AI API enabled.
- Authenticated via Application Default Credentials (ADC).

## Building and Running
This project provides a `Makefile` for common operations.

### Run Locally (Simulation)
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
make run
```

### Build Container Images
```bash
# Build the Go Orchestrator and Python Worker
gcloud builds submit --config cloudbuild.yaml .
```

### Testing
```bash
make test
```

## Project Structure
- `cmd/sandbox/`: Main entry point.
- `internal/ai/`: Gemini 3.1 client wrappers and model routing.
- `internal/orchestrator/`: Tiered Discovery logic and state management.
- `internal/python/`: GKE Sandbox runner (`gke.go`) and local simulation.
- `internal/data/`: SEC 10-K simulator and context generation.
- `deploy/worker/`: Dockerfile for the isolated Python execution environment.
